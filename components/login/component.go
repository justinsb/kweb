package login

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/templates/scopes"
	"k8s.io/klog/v2"
)

type Component struct {
	providers map[string]components.AuthenticationProvider
}

func NewComponent() (*Component, error) {
	return &Component{
		providers: make(map[string]components.AuthenticationProvider),
	}, nil
}

func GetComponent(ctx context.Context) *Component {
	var component *Component
	components.GetComponent(ctx, &component)
	return component
}

// RegisterProvider registers an authentication method with our login system
func (c *Component) RegisterProvider(provider components.AuthenticationProvider) {
	providerID := provider.ProviderID()
	if c.providers[providerID] != nil {
		klog.Fatalf("provider %q already registered", providerID)
	}
	c.providers[providerID] = provider
}

func (c *Component) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/_login/logout", s.ServeHTTP(c.Logout))
	for _, provider := range c.providers {
		provider := provider
		id := provider.ProviderID()
		fn := func(ctx context.Context, req *components.Request) (components.Response, error) {
			return c.StartOAuth2Login(ctx, req, provider)
		}
		mux.HandleFunc("/_login/oauth2/"+id, s.ServeHTTP(fn))
		mux.HandleFunc("/_login/oauth2-callback/"+id, s.ServeHTTP(c.OAuthCallback))
	}

	return nil
}

func (c *Component) AddToScope(ctx context.Context, scope *scopes.Scope) {
	m := map[string]any{
		"logoutURL": "/_login/logout",
	}
	// TODO: which is default?
	for _, provider := range c.providers {
		m["loginURL"] = "/_login/oauth2/" + provider.ProviderID()
		break
	}

	scope.Values["login"] = scopes.Value{Value: m}
}
