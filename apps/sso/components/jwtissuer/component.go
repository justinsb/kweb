package jwtissuer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/components/keystore"
	"github.com/justinsb/kweb/components/users"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"golang.org/x/oauth2/jws"
	"k8s.io/klog/v2"
)

const CookieNameJWT = "auth-token"

type JWTIssuerComponent struct {
	Keys keystore.KeySet

	Issuer   string
	Audience string

	CookieDomain string
}

var _ components.RequestFilter = &JWTIssuerComponent{}

func (c *JWTIssuerComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	log := klog.FromContext(ctx)

	jwtCookie, err := req.Cookie(CookieNameJWT)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}

	jwtCookieValue := ""
	if jwtCookie != nil {
		jwtCookieValue = jwtCookie.Value
	}
	jwtCookieValue = strings.TrimPrefix(jwtCookieValue, "Bearer ")

	user := users.GetUser(ctx)

	setJWT := false
	if jwtCookieValue != "" {
		minTTL := 5 * time.Minute
		if reason := c.jwtIsExpiredOrInvalid(ctx, jwtCookieValue, user, minTTL); reason != "" {
			log.Info("jwt is not valid/is expiring soon; will replace", "reason", reason)
			// If the JWT is bad, we should either replace or remove it
			setJWT = true
		}
	}

	if user != nil && jwtCookieValue == "" {
		setJWT = true
	}

	if setJWT {
		setCookie := http.Cookie{
			Name:     CookieNameJWT,
			HttpOnly: true,
			Secure:   true,
			Path:     "/", // Otherwise cookie is filtered
		}

		if user != nil {
			scopes := []string{"sso"}
			jwtExpiration := time.Hour // while we're developing
			// The istio jwt rule is not very easy to work with,
			// when the cookie has expired it blocks everything,
			// and it's surprisingly tricky to allowlist just one app
			cookieExpiration := jwtExpiration - 5*time.Minute

			setCookie.Expires = time.Now().Add(cookieExpiration)

			token, err := c.buildJWTToken(user.GetMetadata().GetName(), scopes, jwtExpiration)
			if err != nil {
				return nil, fmt.Errorf("error building JWT token: %w", err)
			}
			setCookie.Value = token.TokenType + " " + token.AccessToken
		} else {
			setCookie.Expires = time.Unix(0, 0)
		}

		if c.CookieDomain != "" {
			setCookie.Domain = c.CookieDomain
		}

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

	return next(ctx, req)
}

func (c *JWTIssuerComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/.well-known/openid-configuration", s.ServeHTTP(c.ServeOpenIDConfiguration))
	mux.HandleFunc("/.oidc/jwks", s.ServeHTTP(c.ServeJWKS))
	return nil
}

func (c *JWTIssuerComponent) Key() string {
	return "jwtissuer"
}

func (c *JWTIssuerComponent) ScopeValues() any {
	return nil
}

func (c *JWTIssuerComponent) jwtIsExpiredOrInvalid(ctx context.Context, jwt string, user *userapi.User, minTTL time.Duration) string {
	log := klog.FromContext(ctx)
	tokens := strings.Split(jwt, ".")
	if len(tokens) != 3 {
		log.Info("jwt did not have 3 components; treating as invalid")
		return "corrupt"
	}
	b1, err := base64.RawStdEncoding.DecodeString(tokens[0])
	if err != nil {
		log.Info("jwt component 1 could not be base64 decoded; treating as invalid")
		return "corrupt"
	}
	v1 := make(map[string]interface{})
	if err := json.Unmarshal(b1, &v1); err != nil {
		log.Info("jwt component 1 could not be unmarshaled as json; treating as invalid")
		return "corrupt"
	}

	b2, err := base64.RawStdEncoding.DecodeString(tokens[1])
	if err != nil {
		log.Info("jwt claims could not be base64 decoded; treating as invalid")
		return "corrupt"
	}
	var claims jws.ClaimSet
	if err := json.Unmarshal(b2, &claims); err != nil {
		log.Info("jwt claims could not be unmarshaled as json; treating as invalid")
		return "corrupt"
	}
	log.Info("jwt", "component1", v1, "claims", claims)

	if claims.Exp == 0 {
		log.Info("jwt did not have expiry; treating as invalid")
		return "expired"
	}
	expiresAtTime := time.Unix(claims.Exp, 0)
	ttl := time.Until(expiresAtTime)
	if ttl < minTTL {
		log.Info("jwt expiring soon; marking JWT as invalid")
		return "expiring"
	}
	return ""
}
