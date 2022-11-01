package oidclogin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users/pb"
	"k8s.io/klog/v2"
)

const CookieNameJWT = "auth-token"

type oidcAuthenticator struct {
	opt Options

	mutex        sync.Mutex
	lazyVerifier *oidc.IDTokenVerifier
}

type Options struct {
	Issuer   string
	Audience string
}

func newOIDCAuthenticator(opt Options) *oidcAuthenticator {
	// We lazy-init so we can read our own tokens:
	// verifier fetches /.well-known/openid-configuration, if we're serving that we can't start up

	m := &oidcAuthenticator{
		opt: opt,
	}

	return m
}

func (a *oidcAuthenticator) verifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.lazyVerifier != nil {
		return a.lazyVerifier, nil
	}

	provider, err := oidc.NewProvider(ctx, a.opt.Issuer)
	if err != nil {
		return nil, fmt.Errorf("error building OIDC provider for %q: %w", a.opt.Issuer, err)
	}

	var oidcConfig oidc.Config
	oidcConfig.ClientID = a.opt.Audience

	verifier := provider.Verifier(&oidcConfig)
	a.lazyVerifier = verifier
	return verifier, nil
}

func (c *OIDCLoginComponent) userFromJWTToken(ctx context.Context, req *components.Request, authentictor *oidcAuthenticator) (*pb.User, error) {
	jwtCookie, err := req.Cookie(CookieNameJWT)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}
	if jwtCookie == nil || jwtCookie.Value == "" {
		return nil, nil
	}
	tokenInfo, ok := authentictor.tryAuthenticateJWT(ctx, jwtCookie.Value)
	if !ok {
		return nil, nil
	}

	var rawClaims map[string]interface{}
	tokenInfo.token.Claims(&rawClaims)
	klog.Infof("tokenInfo is %#v", tokenInfo.token)
	klog.Infof("rawClaims is %#v", rawClaims)

	userID := tokenInfo.token.Subject
	if userID != "" {
		user, err := c.userComponent.LoadUser(ctx, userID)
		if err != nil {
			klog.Warningf("failed to load user: %v", err)
			return nil, err
		}
		if user == nil {
			klog.Warningf("user %q was in JWT but was not found", userID)
			return nil, nil
		}
		return user, nil
	}
	return nil, nil
}

type tokenInfo struct {
	token  *oidc.IDToken
	scopes []string
}

// tryAuthenticateJWT checks the token to see if it is valid
func (c *oidcAuthenticator) tryAuthenticateJWT(ctx context.Context, jwt string) (*tokenInfo, bool) {
	if jwt == "" {
		return nil, false
	}

	verifier, err := c.verifier(ctx)
	if err != nil {
		klog.Warningf("error building OIDC verifier: %v", err)
		return nil, false
	}

	jwt = strings.TrimPrefix(jwt, "Bearer ")
	// Parse and verify the jwt
	token, err := verifier.Verify(ctx, jwt)
	if err != nil {
		klog.Warningf("failed to verify token: %v", err)
		return nil, false
	}

	// Extract additional claims
	var claims struct {
		// Scope holds a space-separated list of scopes
		// https://datatracker.ietf.org/doc/html/rfc8693#section-4.2
		Scope string `json:"scope"`
	}
	if err := token.Claims(&claims); err != nil {
		klog.Warningf("failed to extract JWT claims: %v", err)
		return nil, false
	}

	return &tokenInfo{token: token, scopes: strings.Split(claims.Scope, " ")}, true
}
