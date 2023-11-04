package healthcheck

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/templates/scopes"
)

func NewHealthcheckComponent() components.Component {
	return &HealthcheckComponent{}
}

type HealthcheckComponent struct {
}

func (c *HealthcheckComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/healthz", s.ServeHTTP(c.Healthz))
	return nil
}
func (c *HealthcheckComponent) AddToScope(ctx context.Context, scope *scopes.Scope) {
}

func (p *HealthcheckComponent) Healthz(ctx context.Context, req *components.Request) (components.Response, error) {
	html := "ok"
	response := components.SimpleResponse{
		Body: []byte(html),
	}
	return response, nil
}
