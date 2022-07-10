package loginwithgithub

import (
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/login"
	// "github.com/justinsb/kweb/components/login/providers"
)

type Component struct {
	common login.Component
}

func NewComponent(userMapper login.UserMapper, provider *GithubProvider) (*Component, error) {
	return &Component{
		common: login.Component{
			UserMapper: userMapper,
			Provider:   provider,
		},
	}, nil
}

func (c *Component) RegisterHandlers(s *components.Server, mux *http.ServeMux) {
	mux.HandleFunc("/_login/logout", s.ServeHTTP(c.common.Logout))
	mux.HandleFunc("/_login/oauth2/github", s.ServeHTTP(c.common.OAuthStart))
	mux.HandleFunc("/_login/oauth2-callback/github", s.ServeHTTP(c.common.OAuthCallback))

	// Temporary endpoint until we can do more (e.g. UI templating)
	mux.HandleFunc("/info", s.ServeHTTP(c.common.DebugInfo))
}
