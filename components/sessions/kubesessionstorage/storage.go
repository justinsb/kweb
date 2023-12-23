package kubesessionstorage

import (
	"context"
	"encoding/base32"
	"strings"

	cryptorand "crypto/rand"

	"github.com/justinsb/kweb/components/kube/kubeclient"
	"github.com/justinsb/kweb/components/sessions"
	"github.com/justinsb/kweb/components/sessions/kubesessionstorage/api"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type KubeSessionStorage struct {
	kube *kubeclient.Client
}

var _ sessions.Storage = &KubeSessionStorage{}

func NewKubeSessionStorage(kube *kubeclient.Client) *KubeSessionStorage {
	return &KubeSessionStorage{
		kube: kube,
	}
}

func (s *KubeSessionStorage) LookupSession(ctx context.Context, sessionID string) (*sessions.Session, error) {
	if sessionID == "" {
		return nil, nil
	}

	obj := api.Session{}
	key := types.NamespacedName{
		Namespace: "sessions",
		Name:      sessionID,
	}
	if err := s.kube.Uncached().Get(ctx, key, &obj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return sessions.Decode(sessionID, obj.Spec.Data)
}

func (s *KubeSessionStorage) WriteSession(ctx context.Context, session *sessions.Session) error {
	sessionID := session.ID
	create := false
	if sessionID == "" {
		// A new session
		create = true
		sessionID = GenerateSessionID()
		session.ID = sessionID
	}

	b, err := sessions.Encode(session)
	if err != nil {
		return err
	}

	klog.Infof("storing session %q", sessionID)
	if create {
		obj := api.Session{}
		obj.Spec.Data = b
		obj.Namespace = "sessions"
		obj.Name = sessionID
		if err := s.kube.Uncached().Create(ctx, &obj); err != nil {
			return err
		}
	} else {
		obj := api.Session{}
		key := types.NamespacedName{
			Namespace: "sessions",
			Name:      sessionID,
		}
		if err := s.kube.Uncached().Get(ctx, key, &obj); err != nil {
			return err
		}
		obj.Spec.Data = b
		if err := s.kube.Uncached().Update(ctx, &obj); err != nil {
			return err
		}
	}

	return nil
}

func GenerateSessionID() string {
	b := make([]byte, 32, 32)
	if _, err := cryptorand.Read(b); err != nil {
		klog.Fatalf("error building session id: %v", err)
	}
	sessionID := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	sessionID = strings.ToLower(sessionID)
	return sessionID
}
