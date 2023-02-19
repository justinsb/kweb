package loginwithgoogle

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/login"
	"github.com/justinsb/kweb/templates/scopes"
	// "github.com/justinsb/kweb/components/login/providers"
)

type Component struct {
	common login.Component
}

func NewComponent(userMapper login.UserMapper, provider *GoogleProvider) (*Component, error) {
	return &Component{
		common: login.Component{
			UserMapper: userMapper,
			Provider:   provider,
		},
	}, nil
}

func (c *Component) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/_login/logout", s.ServeHTTP(c.common.Logout))
	mux.HandleFunc("/_login/oauth2/google", s.ServeHTTP(c.common.OAuthStart))
	mux.HandleFunc("/_login/oauth2-callback/google", s.ServeHTTP(c.common.OAuthCallback))

	return nil
}

func (c *Component) AddToScope(ctx context.Context, scope *scopes.Scope) {
	m := map[string]any{
		"logoutURL": "/_login/logout",
		"loginURL":  "/_login/oauth2/google",
	}
	scope.Values["login"] = scopes.Value{Value: m}
}
