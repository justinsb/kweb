package keystore

import (
	"context"
	"crypto"

	"github.com/justinsb/kweb/components/keystore/pb"
)

type KeyStore interface {
	KeySet(ctx context.Context, keyname string, keyType pb.KeyType) (KeySet, error)
}

type KeySet interface {
	AuthenticateAndDecrypt(ciphertext []byte) ([]byte, error)
	Encrypt(plaintext []byte) ([]byte, error)

	ActiveKey() (Key, error)
	AllVersions() ([]Key, error)
}

type Key interface {
	KeyType() pb.KeyType
	PublicKey() (crypto.PublicKey, error)
	KeyID() int32

	Signer() (crypto.Signer, error)
}

type internalKey interface {
	Data() *pb.KeyData
	Key
}
