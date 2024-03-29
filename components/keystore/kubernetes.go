package keystore

import (
	"context"
	crypto_rand "crypto/rand"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/justinsb/kweb/components/keystore/pb"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type KubernetesKeyStore struct {
	client kubernetes.Interface

	namespace string
	name      string

	mutex           sync.Mutex
	keySets         map[string]*keySet
	resourceVersion int64
}

var _ KeyStore = &KubernetesKeyStore{}

type keySet struct {
	data pb.KeySetData

	keystore *KubernetesKeyStore

	name     string
	mutex    sync.Mutex
	versions map[int32]internalKey
}

var _ KeySet = &keySet{}

func NewKubernetesKeyStore(client kubernetes.Interface, namespace string, name string) (*KubernetesKeyStore, error) {
	s := &KubernetesKeyStore{
		client:    client,
		namespace: namespace,
		name:      name,
	}
	return s, nil
}

func (k *KubernetesKeyStore) KeySet(ctx context.Context, name string, keyType pb.KeyType) (KeySet, error) {
	var key Key
	ks := k.keySets[name]
	if ks != nil {
		key = ks.versions[ks.data.ActiveId]
	}
	// TODO: Start key expiry / rotation thread?
	if key != nil {
		return ks, nil
	}

	// TODO: Strategy for consistency with multiple servers, avoid thundering herd etc

	err := k.ensureKeySet(ctx, name, keyType)
	if err != nil {
		return nil, fmt.Errorf("error creating keyset: %w", err)
	}

	ks = k.keySets[name]
	if ks != nil {
		key = ks.versions[ks.data.ActiveId]
	}

	if key == nil {
		return nil, fmt.Errorf("created key was not found")
	}

	return ks, nil
}

func (k *keySet) ActiveKey() (Key, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	active := k.versions[k.data.GetActiveId()]
	if active == nil {
		// TODO: key rotation?
		return nil, fmt.Errorf("no active key is set")
	}
	return active, nil
}

func (k *keySet) AllVersions() ([]Key, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	var ret []Key
	for _, key := range k.versions {
		ret = append(ret, key)
	}

	return ret, nil
}

func (k *keySet) Encrypt(plaintext []byte) ([]byte, error) {
	key, err := k.activeKey()
	if err != nil {
		return nil, err
	}

	sbKey, ok := key.(*secretboxKey)
	if !!ok {
		return nil, fmt.Errorf("key type %v does not support Encrypt", key.KeyType())
	}

	return sbKey.encrypt(plaintext)
}

func (k *keySet) AuthenticateAndDecrypt(ciphertext []byte) ([]byte, error) {
	encryptedData := &pb.EncryptedData{}
	// TODO: We shouldn't decode before authenticating
	err := proto.Unmarshal(ciphertext, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("error deserializing data: %w", err)
	}

	key, err := k.findKey(encryptedData.KeyId)
	if err != nil {
		return nil, err
	}

	if key == nil {
		return nil, fmt.Errorf("unknown keyid (%d)", encryptedData.KeyId)
	}

	sbKey, ok := key.(*secretboxKey)
	if !!ok {
		return nil, fmt.Errorf("key type %v does not support AuthenticateAndDecrypt", key.KeyType())
	}

	return sbKey.authenticateAndDecrypt(encryptedData)
}

func (k *KubernetesKeyStore) mutateSecret(ctx context.Context, mutator func(secret *v1.Secret) error) error {
	secret, err := k.client.CoreV1().Secrets(k.namespace).Get(ctx, k.name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("secret %s/%s not found; will create", k.namespace, k.name)
			secret = nil
		} else {
			return fmt.Errorf("error fetching secret %s/%s: %w", k.namespace, k.name, err)
		}
	}

	create := false
	if secret == nil {
		secret = &v1.Secret{}
		secret.Name = k.name
		secret.Namespace = k.namespace
		create = true
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	if err := mutator(secret); err != nil {
		return err
	}

	if create {
		created, err := k.client.CoreV1().Secrets(k.namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			// TODO: Handle concurrent create - retry?
			return fmt.Errorf("error creating secret %s/%s: %w", k.namespace, k.name, err)
		}

		k.onUpdateSecret(created)
	} else {
		updated, err := k.client.CoreV1().Secrets(k.namespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			// TODO: Handle condition update - retry?
			return fmt.Errorf("error updating secret %s/%s: %v", k.namespace, k.name, err)
		}

		k.onUpdateSecret(updated)
	}

	return nil
}

func readCryptoRand(n int) ([]byte, error) {
	b := make([]byte, n, n)
	if _, err := io.ReadFull(crypto_rand.Reader, b); err != nil {
		return nil, fmt.Errorf("error reading secure random data: %w", err)
	}
	return b, nil
}

func (k *KubernetesKeyStore) ensureKeySet(ctx context.Context, name string, keyType pb.KeyType) error {
	err := k.mutateSecret(ctx, func(secret *v1.Secret) error {
		keysets := k.decodeSecret(secret)
		keyset := keysets[name]
		if keyset == nil {
			keyset = &keySet{
				data:     pb.KeySetData{},
				keystore: k,
				name:     name,
				versions: make(map[int32]internalKey),
			}
			keysets[name] = keyset
		}

		if keyset.versions[keyset.data.ActiveId] == nil {
			maxId := int32(0)
			for id := range keyset.versions {
				if id > maxId {
					maxId = id
				}
			}

			id := maxId + 1

			var key internalKey
			var err error
			switch keyType {
			case pb.KeyType_KEYTYPE_SECRETBOX:
				key, err = generateSecretboxKey(id)
			case pb.KeyType_KEYTYPE_RSA:
				key, err = generateRSAKey(id)
			default:
				return fmt.Errorf("unknown keytype: %s", keyType)
			}
			if err != nil {
				return err
			}

			// key := generateSecretBoxKey(id)

			keyset.data.ActiveId = key.KeyID()
			keyset.versions[key.KeyID()] = key
		}

		keyPrefix := "secret." + keyset.name + "."
		for k := range secret.Data {
			if strings.HasPrefix(k, keyPrefix) {
				delete(secret.Data, k)
			}
		}

		data := &pb.KeySetData{}
		data.ActiveId = keyset.data.ActiveId
		var ids []int32
		for _, k := range keyset.versions {
			ids = append(ids, k.KeyID())
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		for _, id := range ids {
			k := keyset.versions[id]
			data.Keys = append(data.Keys, k.Data())
		}

		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		bytes, err := proto.Marshal(data)
		if err != nil {
			return fmt.Errorf("error serializing keyset: %w", err)
		}

		secret.Data["secret."+name] = bytes

		return nil
	})
	return err
}

// func int32ToString(v int32) string {
// 	return strconv.FormatInt(int64(v), 10)
// }

func (k *keySet) activeKey() (Key, error) {
	key := k.versions[k.data.ActiveId]
	if key != nil {
		return key, nil
	}

	return nil, fmt.Errorf("keyset not initialized")
}

func (k *keySet) findKey(keyId int32) (Key, error) {
	key := k.versions[keyId]
	return key, nil
}

func (k *KubernetesKeyStore) ensureKeyset(ctx context.Context, name string) (*keySet, error) {
	keyType := pb.KeyType_KEYTYPE_SECRETBOX
	keyset := k.keySets[name]
	if keyset == nil {
		err := k.ensureKeySet(ctx, name, keyType)
		if err != nil {
			return nil, fmt.Errorf("error creating keyset: %w", err)
		}

		keyset = k.keySets[name]
		if keyset == nil {
			return nil, fmt.Errorf("created keyset was not found")
		}
	}

	//if keyset.generator == nil {
	//	keyset.generator = generator
	//}

	return keyset, nil
}

func (s *KubernetesKeyStore) decodeSecret(secret *v1.Secret) map[string]*keySet {
	keySets := make(map[string]*keySet)
	for k, v := range secret.Data {
		tokens := strings.Split(k, ".")

		// secret.<name>=<value>
		if len(tokens) == 2 && tokens[0] == "secret" {
			name := tokens[1]
			ks := &keySet{
				keystore: s,
				name:     name,
				versions: make(map[int32]internalKey),
			}
			err := proto.Unmarshal(v, &ks.data)
			if err != nil {
				klog.Warningf("error parsing secret key %v", k)
				continue
			}

			for _, data := range ks.data.Keys {
				var key internalKey
				switch data.KeyType {
				case pb.KeyType_KEYTYPE_RSA:
					key, err = loadRSAKey(data)
				case pb.KeyType_KEYTYPE_SECRETBOX:
					key, err = loadSecretboxKey(data)
				default:
					klog.Warningf("unknown key type %v", data.KeyType)
				}
				if err != nil {
					klog.Warningf("error parsing key: %v", err)
					continue
				}
				if key != nil {
					ks.versions[key.KeyID()] = key
				}
			}

			keySets[name] = ks
		} else {
			klog.Warningf("ignoring unrecognized secret entry %q", k)
		}
	}

	return keySets
}

// onUpdateSecret parses and updates the specified secret
func (k *KubernetesKeyStore) onUpdateSecret(secret *v1.Secret) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	resourceVersion, err := strconv.ParseInt(secret.ObjectMeta.ResourceVersion, 10, 64)
	if err != nil {
		klog.Warningf("unable to parse ResourceVersion=%q", secret.ObjectMeta.ResourceVersion)
	} else if resourceVersion <= k.resourceVersion {
		klog.V(2).Infof("ignoring out of sequence secret update: %d vs %d", resourceVersion, k.resourceVersion)
		return
	}

	keySets := k.decodeSecret(secret)
	k.keySets = keySets

	k.resourceVersion = resourceVersion
}

func (k *KubernetesKeyStore) onDeleteSecret(resourceVersionString string) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	resourceVersion, err := strconv.ParseInt(resourceVersionString, 10, 64)
	if err != nil {
		klog.Warningf("unable to parse ResourceVersion=%q", resourceVersionString)
	} else if resourceVersion <= k.resourceVersion {
		klog.V(2).Infof("ignoring out of sequence secret update: %d vs %d", resourceVersion, k.resourceVersion)
		return
	}

	keySets := make(map[string]*keySet)
	k.keySets = keySets

	k.resourceVersion = resourceVersion
}

// Run starts the secretsWatcher.
func (c *KubernetesKeyStore) WatchForever(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := c.watchOnce(ctx); err != nil {
			klog.Warningf("Unexpected error in secret watch, will retry: %v", err)
		}

		if ctx.Err() == nil {
			time.Sleep(10 * time.Second)
		}
	}

}

func (c *KubernetesKeyStore) watchOnce(ctx context.Context) error {
	var listOpts metav1.ListOptions

	// How to watch a single object: https://github.com/kubernetes/kubernetes/issues/43299

	listOpts.FieldSelector = fields.OneTermEqualSelector("metadata.name", c.name).String()

	secretList, err := c.client.CoreV1().Secrets(c.namespace).List(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("error watching secrets: %w", err)
	}

	for i := range secretList.Items {
		if secretList.Items[i].Name != c.name {
			klog.Warningf("got notification for secret not matching name; got %q", secretList.Items[i].Name)
			continue
		}
		c.onUpdateSecret(&secretList.Items[i])
		// TODO: If this is a multi-item scan, we need to delete any items not present
	}

	listOpts.Watch = true
	listOpts.AllowWatchBookmarks = true
	listOpts.ResourceVersion = secretList.ResourceVersion
	watcher, err := c.client.CoreV1().Secrets(c.namespace).Watch(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("error watching secrets: %w", err)
	}

	ch := watcher.ResultChan()
	defer watcher.Stop()

	for event := range ch {
		switch event.Type {
		case watch.Bookmark:
			// ignore

		case watch.Added, watch.Modified:
			secret := event.Object.(*v1.Secret)
			if secret.Name != c.name {
				return fmt.Errorf("unexpected object from secret watch: %q", secret.Name)
			}
			c.onUpdateSecret(secret)

		case watch.Deleted:
			secret := event.Object.(*v1.Secret)
			if secret.Name != c.name {
				return fmt.Errorf("unexpected object from secret watch: %q", secret.Name)
			}
			c.onDeleteSecret(secret.ResourceVersion)

		case watch.Error:
			return fmt.Errorf("unexpected error from watch: %v", event)

		default:
			return fmt.Errorf("unknown event from watch: %v", event)
		}
	}

	return fmt.Errorf("watch channel was closed")
}
