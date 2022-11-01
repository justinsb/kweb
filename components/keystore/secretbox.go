package keystore

import (
	"crypto"
	crypto_rand "crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/justinsb/kweb/components/keystore/pb"
	"golang.org/x/crypto/nacl/secretbox"
	"google.golang.org/protobuf/proto"
)

type secretboxKey struct {
	data   *pb.KeyData
	secret []byte
}

var _ Key = &secretboxKey{}

func generateSecretboxKey(id int32) (*secretboxKey, error) {
	secretData, err := readCryptoRand(32)
	if err != nil {
		return nil, fmt.Errorf("error generating secret: %w", err)
	}

	data := &pb.KeyData{
		Id:      id,
		Created: time.Now().Unix(),
		Secret:  secretData,
		KeyType: pb.KeyType_KEYTYPE_SECRETBOX,
	}

	return &secretboxKey{
		data:   data,
		secret: secretData,
	}, nil
}

func loadSecretboxKey(data *pb.KeyData) (*secretboxKey, error) {
	return &secretboxKey{
		data:   data,
		secret: data.GetSecret(),
	}, nil
}

func (k *secretboxKey) Data() *pb.KeyData {
	return k.data
}

func (k *secretboxKey) KeyType() pb.KeyType {
	return pb.KeyType_KEYTYPE_SECRETBOX
}

func (k *secretboxKey) KeyID() int32 {
	return k.data.GetId()
}

func (k *secretboxKey) PublicKey() (crypto.PublicKey, error) {
	return nil, fmt.Errorf("symmetric key does not have PublicKey")
}

func (k *secretboxKey) Signer() (crypto.Signer, error) {
	// I guess it can, actually, but we don't need it currently...
	return nil, fmt.Errorf("symmetric key cannot sign data")
}

func (k *secretboxKey) encrypt(plaintext []byte) ([]byte, error) {
	// From the example in the secretbox docs:
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return nil, fmt.Errorf("error reading random data: %w", err)
	}

	secretKey := k.secret
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("expected 32 byte key, was %d", len(secretKey))
	}

	var secretKeyArray [32]byte
	copy(secretKeyArray[:], secretKey[:32])

	ciphertext := secretbox.Seal(nil, plaintext, &nonce, &secretKeyArray)

	encrypted := &pb.EncryptedData{
		EncryptionMethod: pb.EncryptionMethod_ENCRYPTIONMETHOD_SECRETBOX,
		KeyId:            k.data.GetId(),
		Nonce:            nonce[:],
		Ciphertext:       ciphertext,
	}

	encryptedBytes, err := proto.Marshal(encrypted)
	if err != nil {
		return nil, fmt.Errorf("error serializing data: %v", err)
	}

	return encryptedBytes, nil
}

func (k *secretboxKey) authenticateAndDecrypt(encryptedData *pb.EncryptedData) ([]byte, error) {
	if len(encryptedData.Nonce) != 24 {
		return nil, fmt.Errorf("invalid nonce data")
	}

	secretKey := k.secret
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("expected 32 byte key, was %d", len(secretKey))
	}

	var nonceArray [24]byte
	copy(nonceArray[:], encryptedData.Nonce)

	var secretKeyArray [32]byte
	copy(secretKeyArray[:], secretKey[:32])

	plaintext, ok := secretbox.Open(nil, encryptedData.Ciphertext, &nonceArray, &secretKeyArray)
	if !ok {
		return nil, fmt.Errorf("encrypted data not valid")
	}

	return plaintext, nil
}
