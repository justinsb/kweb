package keystore

import (
	"crypto"
	crypto_rand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/justinsb/kweb/components/keystore/pb"
)

type rsaKey struct {
	data *pb.KeyData

	key *rsa.PrivateKey
}

func generateRSAKey(id int32) (*rsaKey, error) {
	key, err := rsa.GenerateKey(crypto_rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("error from rsa.GenerateKey: %w", err)
	}

	secretData := x509.MarshalPKCS1PrivateKey(key)
	data := &pb.KeyData{
		Id:      id,
		Secret:  secretData,
		Created: time.Now().Unix(),
		KeyType: pb.KeyType_KEYTYPE_RSA,
	}
	return &rsaKey{
		data: data,
		key:  key,
	}, nil
}

func loadRSAKey(data *pb.KeyData) (*rsaKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(data.GetSecret())
	if err != nil {
		return nil, fmt.Errorf("key is corrupt")
	}
	return &rsaKey{
		data: data,
		key:  key,
	}, nil
}

func (k *rsaKey) Data() *pb.KeyData {
	return k.data
}

func (k *rsaKey) KeyType() pb.KeyType {
	return pb.KeyType_KEYTYPE_RSA
}

func (k *rsaKey) PublicKey() (crypto.PublicKey, error) {
	return &k.key.PublicKey, nil
}

func (k *rsaKey) KeyID() int32 {
	return k.data.GetId()
}

func (k *rsaKey) Signer() (crypto.Signer, error) {
	return k.key, nil
}

// func (k *rsaKey) sign(plaintext []byte) ([]byte, error) {
// 	hash := crypto.SHA256
// 	hasher := hash.New()
// 	if _, err := hasher.Write(plaintext); err != nil {
// 		// This shouldn't happen
// 		return nil, fmt.Errorf("error during hashing: %w", err)
// 	}
// 	hashed := hasher.Sum(nil)

// 	signed, err := rsa.SignPKCS1v15(crypto_rand.Reader, k.key, hash, hashed)
// 	if err != nil {
// 		return nil, fmt.Errorf("error signing with key: %w", err)
// 	}
// 	return signed, nil
// }

func (k *rsaKey) checkSignature(payload []byte, signature []byte) error {
	hash := crypto.SHA256
	hasher := hash.New()
	if _, err := hasher.Write(payload); err != nil {
		// This shouldn't happen
		return fmt.Errorf("error during hashing: %w", err)
	}
	hashed := hasher.Sum(nil)

	return rsa.VerifyPKCS1v15(&k.key.PublicKey, hash, hashed, signature)
}
