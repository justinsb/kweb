package certs

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
)

func LoadCertificatesFromFile(p string) ([]*x509.Certificate, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", p, err)
	}
	certs, err := DecodeCertificates(b)
	if err != nil {
		return nil, fmt.Errorf("reading certificates from file %q: %w", p, err)
	}
	return certs, nil
}

func DecodeCertificates(b []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate
	for {
		block, remainder := pem.Decode(b)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parsing certificate: %w", err)
			}
			certificates = append(certificates, cert)
		} else {
			return nil, fmt.Errorf("unexpected PEM block type %q", block.Type)
		}
		b = remainder
	}
	return certificates, nil
}

func DecodeCertificate(b []byte) (*x509.Certificate, error) {
	certs, err := DecodeCertificates(b)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}
	if len(certs) > 1 {
		return nil, fmt.Errorf("multiple certificates found")
	}
	return certs[0], nil
}

func LoadCertificate(p string) (*x509.Certificate, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", p, err)
	}

	cert, err := DecodeCertificate(b)
	if err != nil {
		return nil, fmt.Errorf("reading certificate from file %q: %w", p, err)
	}
	return cert, nil
}

func LoadPrivateKey(p string) (crypto.PrivateKey, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", p, err)
	}
	key, err := DecodePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("reading key from file %q: %w", p, err)
	}
	return key, nil
}

func DecodePrivateKey(b []byte) (crypto.PrivateKey, error) {
	var keys []crypto.PrivateKey
	for {
		block, remainder := pem.Decode(b)
		if block == nil {
			break
		}
		if block.Type == "PRIVATE KEY" {
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parsing private key: %w", err)
			}
			privateKey, ok := key.(crypto.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("unexpected type for private key, got %T", key)
			}
			keys = append(keys, privateKey)
		} else {
			return nil, fmt.Errorf("unexpected PEM block type %q", block.Type)
		}
		b = remainder
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no key found")
	}
	if len(keys) > 1 {
		return nil, fmt.Errorf("multiple keys found")
	}
	return keys[0], nil
}

func WriteCertificate(p string, cert *x509.Certificate) error {
	b, err := EncodeCertificate(cert)
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, b, 0644); err != nil {
		return fmt.Errorf("writing certificate to file %q: %w", p, err)
	}
	return nil
}

func EncodeCertificates(certs []*x509.Certificate) ([]byte, error) {
	var buf bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return nil, fmt.Errorf("failed to encode cert to pem")
		}
	}
	return buf.Bytes(), nil
}

func EncodeCertificate(cert *x509.Certificate) ([]byte, error) {
	derBytes := cert.Raw
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return nil, fmt.Errorf("failed to encode cert to pem")
	}
	return pemCert, nil
}

func WritePrivateKey(p string, privateKey crypto.PrivateKey) error {
	b, err := EncodePrivateKey(privateKey)
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		return fmt.Errorf("writing private key file %q: %w", p, err)
	}

	return nil
}

func EncodePrivateKey(privateKey crypto.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("encoding private key: %w", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		return nil, fmt.Errorf("failed to encode key to pem")
	}
	return pemKey, nil
}

func CreateCertificate(template x509.Certificate, signerCert *x509.Certificate, signerKey crypto.PrivateKey) (*x509.Certificate, crypto.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("generating serial number: %w", err)
	}
	template.SerialNumber = serialNumber

	if signerCert == nil {
		// Self-signed
		signerCert = &template
		signerKey = privateKey
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, signerCert, &privateKey.PublicKey, signerKey)
	if err != nil {
		return nil, nil, fmt.Errorf("creating certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing certificate: %w", err)
	}

	return cert, privateKey, nil
}
