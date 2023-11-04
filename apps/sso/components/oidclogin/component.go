package oidclogin

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/apps/sso/pkg/oidc"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users"
	"github.com/justinsb/kweb/templates/scopes"
)

type OIDCLoginComponent struct {
	oidcAuthenticator *oidc.Authenticator // todo: multiple?
}

func NewOIDCLoginComponent(ctx context.Context, oidcAuthentiator *oidc.Authenticator, userComponent *users.UserComponent) *OIDCLoginComponent {
	return &OIDCLoginComponent{
		oidcAuthenticator: oidcAuthentiator,
	}
}

var _ components.RequestFilter = &OIDCLoginComponent{}

func (c *OIDCLoginComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	userInfo, err := c.userFromJWTToken(ctx, req, c.oidcAuthenticator)
	if err != nil {
		return nil, err
	}
	users.SetUser(ctx, userInfo)

	return next(ctx, req)
}

func (c *OIDCLoginComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *OIDCLoginComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
}
