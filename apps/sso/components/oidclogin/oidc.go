package oidclogin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinsb/kweb/apps/sso/pkg/oidc"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users/pb"
)

const CookieNameJWT = "auth-token"

func (c *OIDCLoginComponent) userFromJWTToken(ctx context.Context, req *components.Request, authentictor *oidc.Authenticator) (*pb.User, error) {
	jwtCookie, err := req.Cookie(CookieNameJWT)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}
	if jwtCookie == nil || jwtCookie.Value == "" {
		return nil, nil
	}
	return c.oidcAuthenticator.UserFromJWT(ctx, jwtCookie.Value)
}
