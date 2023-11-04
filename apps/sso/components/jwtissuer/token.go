package jwtissuer

import (
	"crypto"
	crypto_rand "crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/justinsb/kweb/components/keystore/pb"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

// buildToken creates a new JWT token
func (c *JWTIssuerComponent) buildJWTToken(userID string, scopes []string, expiration time.Duration) (*oauth2.Token, error) {
	var claims jws.ClaimSet
	// email address of the client_id of the application making the access token request
	claims.Iss = c.oidcAuthenticator.GetIssuer()
	// space-delimited list of the permissions the application requests
	claims.Scope = strings.Join(scopes, " ")

	// descriptor of the intended target of the assertion (Optional).
	claims.Aud = c.oidcAuthenticator.GetAudience()
	now := time.Now().Unix()

	// the time the assertion was issued (seconds since Unix epoch)
	claims.Iat = now
	// the expiration time of the assertion (seconds since Unix epoch)
	claims.Exp = now + int64(expiration.Seconds())

	claims.Sub = userID

	key, err := c.keys.ActiveKey()
	if err != nil {
		return nil, fmt.Errorf("unable to get active key: %w", err)
	}

	var header jws.Header
	header.KeyID = strconv.FormatInt(int64(key.KeyID()), 10)
	header.Typ = "JWT"

	var hashAlg crypto.Hash
	switch key.KeyType() {
	case pb.KeyType_KEYTYPE_RSA:
		hashAlg = crypto.SHA256
		header.Algorithm = "RS256"

	default:
		return nil, fmt.Errorf("unhandled key type %v", key.KeyType())
	}

	signer, err := key.Signer()
	if err != nil {
		return nil, fmt.Errorf("error building signer: %w", err)
	}
	jwsSigner := func(data []byte) ([]byte, error) {
		hasher := hashAlg.New()
		if _, err := hasher.Write(data); err != nil {
			return nil, fmt.Errorf("error hashing data: %w", err)
		}
		hashed := hasher.Sum(nil)
		return signer.Sign(crypto_rand.Reader, hashed, hashAlg)
	}
	encoded, err := jws.EncodeWithSigner(&header, &claims, jwsSigner)
	if err != nil {
		return nil, fmt.Errorf("error building JWT token: %w", err)
	}
	token := &oauth2.Token{
		AccessToken: encoded,
		TokenType:   "Bearer",
	}
	return token, nil
}
