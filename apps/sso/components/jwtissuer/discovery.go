package jwtissuer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/justinsb/kweb/components"
	"gopkg.in/square/go-jose.v2"
)

type discovery struct {
	Issuer  string `json:"issuer"`
	JwksURI string `json:"jwks_uri"`
}

func (c *JWTIssuerComponent) ServeOpenIDConfiguration(ctx context.Context, req *components.Request) (components.Response, error) {
	response := &discovery{}

	response.Issuer = c.Issuer
	response.JwksURI = strings.TrimSuffix(c.Issuer, "/") + "/.oidc/jwks"

	return &components.JSONResponse{Object: response}, nil
}

func (c *JWTIssuerComponent) ServeJWKS(ctx context.Context, req *components.Request) (components.Response, error) {
	response := &jose.JSONWebKeySet{}
	keys, err := c.Keys.AllVersions()
	if err != nil {
		return nil, fmt.Errorf("error listing keys: %w", err)
	}
	for _, key := range keys {
		publicKey, err := key.PublicKey()
		if err != nil {
			return nil, fmt.Errorf("unable to get public key: %w", err)
		}
		keyID := key.KeyID()
		response.Keys = append(response.Keys, jose.JSONWebKey{
			Key:   publicKey,
			KeyID: strconv.FormatInt(int64(keyID), 10),
		})
	}

	return &components.JSONResponse{Object: response}, nil
}
