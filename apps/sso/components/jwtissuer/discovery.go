package jwtissuer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/justinsb/kweb/components"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"gopkg.in/square/go-jose.v2"
	"k8s.io/klog/v2"
)

type discovery struct {
	Issuer           string `json:"issuer"`
	JwksURI          string `json:"jwks_uri"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
}

func (c *JWTIssuerComponent) ServeOpenIDConfiguration(ctx context.Context, req *components.Request) (components.Response, error) {
	response := &discovery{}

	issuer := c.oidcAuthenticator.GetIssuer()
	response.Issuer = issuer
	response.JwksURI = strings.TrimSuffix(issuer, "/") + "/.oidc/jwks"
	response.UserInfoEndpoint = strings.TrimSuffix(issuer, "/") + "/.oidc/userinfo"

	return &components.JSONResponse{Object: response}, nil
}

func (c *JWTIssuerComponent) ServeJWKS(ctx context.Context, req *components.Request) (components.Response, error) {
	response := &jose.JSONWebKeySet{}
	keys, err := c.keys.AllVersions()
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

type userInfo struct {
	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

func (c *JWTIssuerComponent) ServeUserInfo(ctx context.Context, req *components.Request) (components.Response, error) {
	auth := req.Header.Get("Authorization")

	token := ""
	if strings.HasPrefix(auth, "Bearer ") {
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	if token == "" {
		return components.ErrorResponse(http.StatusUnauthorized), nil
	}

	var user *userapi.User
	if token != "" {
		u, err := c.oidcAuthenticator.UserFromJWT(ctx, token)
		if err != nil {
			klog.Warningf("error get user from jwt: %v", err)
			return components.ErrorResponse(http.StatusInternalServerError), nil
		}
		user = u
	}

	if user != nil {
		email := user.Spec.Email
		name := "" // TODO
		if name == "" {
			name = email
		}
		preferredUsername := "" // TODO
		if preferredUsername == "" {
			preferredUsername = email
		}
		info := &userInfo{
			Name:              name,
			Email:             email,
			PreferredUsername: preferredUsername,
		}

		return &components.JSONResponse{Object: info}, nil
	}

	return components.ErrorResponse(http.StatusUnauthorized), nil
}
