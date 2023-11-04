package oidc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/justinsb/kweb/components/users"
	"k8s.io/klog/v2"
)

type Authenticator struct {
	opt Options

	users *users.UserComponent

	mutex        sync.Mutex
	lazyVerifier *oidc.IDTokenVerifier
}

func NewAuthenticator(users *users.UserComponent, opt Options) *Authenticator {
	// We lazy-init so we can read our own tokens:
	// verifier fetches /.well-known/openid-configuration, if we're serving that we can't start up

	m := &Authenticator{
		opt:   opt,
		users: users,
	}

	return m
}

type Options struct {
	Issuer   string
	Audience string
}

func (a *Authenticator) GetIssuer() string {
	return a.opt.Issuer
}

func (a *Authenticator) GetAudience() string {
	return a.opt.Audience
}

func (a *Authenticator) verifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
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

type tokenInfo struct {
	token  *oidc.IDToken
	scopes []string
}

// tryAuthenticateJWT checks the token to see if it is valid
func (c *Authenticator) tryAuthenticateJWT(ctx context.Context, jwt string) (*tokenInfo, bool) {
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
