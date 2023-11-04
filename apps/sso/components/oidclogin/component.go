package oidclogin

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users"
)

type OIDCLoginComponent struct {
	oidcAuthenticator *oidcAuthenticator // todo: multiple?

	userComponent *users.UserComponent
}

func NewOIDCLoginComponent(ctx context.Context, opt Options, userComponent *users.UserComponent) *OIDCLoginComponent {
	authenticator := newOIDCAuthenticator(opt)
	return &OIDCLoginComponent{
		oidcAuthenticator: authenticator,
		userComponent:     userComponent,
	}
}

var _ components.RequestFilter = &OIDCLoginComponent{}

func (c *OIDCLoginComponent) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	userInfo, err := c.userFromJWTToken(ctx, req, c.oidcAuthenticator)
	if err != nil {
		return nil, err
	}
	if userInfo != nil {
		ctx = users.WithUser(ctx, userInfo)
	}

	return next(ctx, req)
}

func (c *OIDCLoginComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	return nil
}

func (c *OIDCLoginComponent) Key() string {
	return "jwtissuer"
}

func (c *OIDCLoginComponent) ScopeValues() any {
	return nil
}
