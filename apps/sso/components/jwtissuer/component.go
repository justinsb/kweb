package jwtissuer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/components/keystore"
	"github.com/justinsb/kweb/components/users"
	"k8s.io/klog/v2"
)

const CookieNameJWT = "auth-token"

type JWTIssuerComponent struct {
	Keys keystore.KeySet

	Issuer   string
	Audience string
}

var _ components.RequestFilter = &JWTIssuerComponent{}

func (c *JWTIssuerComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	user := users.GetUser(ctx)

	if user != nil {
		jwtCookie, err := req.Cookie(CookieNameJWT)
		if err != nil && err != http.ErrNoCookie {
			return nil, fmt.Errorf("error reading cookie: %w", err)
		}
		if jwtCookie == nil || jwtCookie.Value == "" {
			scopes := []string{"sso"}
			expiration := time.Hour // while we're developing
			token, err := c.buildToken(user.GetMetadata().GetName(), scopes, expiration)
			if err != nil {
				return nil, fmt.Errorf("error building JWT token: %w", err)
			}

			setCookie := http.Cookie{
				Name:     CookieNameJWT,
				Value:    token.TokenType + " " + token.AccessToken,
				Expires:  time.Now().Add(time.Hour * 24 * 365),
				HttpOnly: true,
				Secure:   true,
				Path:     "/", // Otherwise cookie is filtered
			}

			// TODO: Set domain
			// setCookie.Domain = ...

			if !req.BrowserUsingHTTPS() {
				if req.IsLocalhost() {
					klog.Warningf("setting cookie to _not_ be secure, because running on localhost")
					setCookie.Secure = false
				} else {
					klog.Warningf("session invoked but running without TLS (and not on localhost); likely won't work")
				}
			}

			cookies.SetCookie(ctx, setCookie)
		}
	}

	return next(ctx, req)
}

func (c *JWTIssuerComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/.well-known/openid-configuration", s.ServeHTTP(c.ServeOpenIDConfiguration))
	mux.HandleFunc("/.oidc/jwks", s.ServeHTTP(c.ServeJWKS))
	return nil
}
