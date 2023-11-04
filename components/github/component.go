package github

import (
	"context"
	"crypto/rsa"
	"net/http"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/kube/kubeclient"
)

type Component struct {
	kube *kubeclient.Client

	githubAppID string

	appPrivateKey *rsa.PrivateKey
}

func New(kube *kubeclient.Client, githubAppID string, appPrivateKey *rsa.PrivateKey) (*Component, error) {
	return &Component{
		kube:          kube,
		githubAppID:   githubAppID,
		appPrivateKey: appPrivateKey,
	}, nil
}

func (c *Component) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	mux.HandleFunc("/_ghapp/", s.ServeHTTP(c.doEntryPoint))
	return nil
}

func (c *Component) Key() string {
	return "github"
}

func (c *Component) ScopeValues() any {
	return nil
}

var _ components.RequestFilter = &Component{}

func (c *Component) ProcessRequest(ctx context.Context, req *components.Request, next components.RequestFilterChain) (components.Response, error) {
	ctx = context.WithValue(ctx, contextKeyRequest, &requestInfo{component: c})

	return next(ctx, req)
}
