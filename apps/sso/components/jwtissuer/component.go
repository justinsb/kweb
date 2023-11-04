package jwtissuer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinsb/kweb/apps/sso/pkg/oidc"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/cookies"
	"github.com/justinsb/kweb/components/keystore"
	"github.com/justinsb/kweb/components/users"
	userapi "github.com/justinsb/kweb/components/users/pb"
	"github.com/justinsb/kweb/templates/scopes"
	"golang.org/x/oauth2/jws"
	"k8s.io/klog/v2"
)

const CookieNameJWT = "auth-token"

type Options struct {
	CookieDomain string
}

type JWTIssuerComponent struct {
	keys keystore.KeySet

	oidcAuthenticator *oidc.Authenticator
	opts              Options
}

func NewJWTIssuerComponent(keys keystore.KeySet, oidcAuthenticator *oidc.Authenticator, opts Options) *JWTIssuerComponent {
	return &JWTIssuerComponent{
		keys:              keys,
		oidcAuthenticator: oidcAuthenticator,
		opts:              opts,
	}
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
		if user == nil {
			setJWT = true
		} else {
			minTTL := 5 * time.Minute
			if reason := c.jwtIsExpiredOrInvalid(ctx, jwtCookieValue, user, minTTL); reason != "" {
				log.Info("jwt is not valid/is expiring soon; will replace", "reason", reason)
				// If the JWT is bad, we should either replace or remove it
				setJWT = true
			}
		}
	}

	if user != nil && jwtCookieValue == "" {
		setJWT = true
	}

	if setJWT {
		if err := c.setCookie(ctx, req, user); err != nil {
			return nil, err
		}
	}

	redirect := req.Request.FormValue("redirect")
	if redirect != "" {
		req.Session.SetString("redirect", redirect)
	}
	if user != nil {
		if redirect == "" {
			redirect = req.Session.GetString("redirect")
		}
		if redirect != "" {
			return components.RedirectResponse(redirect), nil
		}
	}

	oldUser := user

	response, err := next(ctx, req)
	if err == nil {
		newUser := users.GetUser(ctx)
		if newUser == nil {
			if oldUser != nil {
				// logout
				klog.Infof("user logout detected; clearing cookie")
				if err := c.setCookie(ctx, req, nil); err != nil {
					return nil, err
				}
			}
		} else if oldUser == nil || (oldUser.Metadata.Uid == newUser.Metadata.Uid) {
			// login
			klog.Infof("user login detected; setting cookie")
			if err := c.setCookie(ctx, req, newUser); err != nil {
				return nil, err
			}
		}
	}

	return response, err
}

func (c *JWTIssuerComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/.well-known/openid-configuration", s.ServeHTTP(c.ServeOpenIDConfiguration))
	mux.HandleFunc("/.oidc/jwks", s.ServeHTTP(c.ServeJWKS))
	mux.HandleFunc("/.oidc/userinfo", s.ServeHTTP(c.ServeUserInfo))
	return nil
}

func (c *JWTIssuerComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
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

func (c *JWTIssuerComponent) setCookie(ctx context.Context, req *components.Request, user *userapi.User) error {
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
			return fmt.Errorf("error building JWT token: %w", err)
		}
		setCookie.Value = token.TokenType + " " + token.AccessToken
	} else {
		setCookie.Expires = time.Unix(0, 0)
	}

	if c.opts.CookieDomain != "" {
		setCookie.Domain = c.opts.CookieDomain
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
	return nil
}
