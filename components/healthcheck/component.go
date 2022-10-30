package healthcheck

import (
	"context"
	"net/http"

	"github.com/justinsb/kweb/components"
)

func NewHealthcheckComponent() components.Component {
	return &HealthcheckComponent{}
}

type HealthcheckComponent struct {
}

func (c *HealthcheckComponent) RegisterHandlers(s *components.Server, mux *http.ServeMux) {
	mux.HandleFunc("/healthz", s.ServeHTTP(c.Healthz))
}

func (p *HealthcheckComponent) Healthz(ctx context.Context, req *components.Request) (components.Response, error) {
	html := "ok"
	response := components.SimpleResponse{
		Body: []byte(html),
	}
	return response, nil
}
