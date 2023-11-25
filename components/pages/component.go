package pages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/templates"
	"github.com/justinsb/kweb/templates/scopes"
	"k8s.io/klog/v2"
)

type Options struct {
	Base fs.FS

	ScopeValues []ScopeFunction
}

type ScopeFunction func(ctx context.Context, scope *scopes.Scope)

func (o *Options) InitDefaults(appName string) {
	o.Base = os.DirFS("pages")
}

type Component struct {
	options Options
}

func New(opt Options) *Component {
	return &Component{options: opt}
}

func loadRaw(fs fs.FS, key string) ([]byte, error) {
	f, err := fs.Open(key)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %w", key, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %w", key, err)
	}

	return b, nil
}

func (c *Component) RegisterHandlers(s *components.Server, mux *http.ServeMux) error {
	if err := c.addHandlersFromDir(s, mux, "."); err != nil {
		return err
	}
	return nil
}

func (c *Component) AddToScope(ctx context.Context, scope *scopes.Scope) {
	for _, scopeFunc := range c.options.ScopeValues {
		scopeFunc(ctx, scope)
	}
}

func (c *Component) addHandlersFromDir(s *components.Server, mux *http.ServeMux, p string) error {
	entries, err := fs.ReadDir(c.options.Base, p)
	if err != nil {
		return fmt.Errorf("error from ReadDir(%q): %w", p, err)
	}

	for _, entry := range entries {
		name := path.Join(p, entry.Name())
		if err := c.addHandlers(s, mux, name, entry); err != nil {
			return err
		}
	}

	return nil
}

func (c *Component) addHandlers(s *components.Server, mux *http.ServeMux, p string, info fs.DirEntry) error {
	if !info.IsDir() {
		templateData, err := loadRaw(c.options.Base, p)
		if err != nil {
			return fmt.Errorf("error reading %q: %w", p, err)
		}

		template := templates.Template{
			Data: []byte(templateData),
		}

		endpoint := &TemplateEndpoint{template: template, server: s}
		serveOn := "/" + p
		// Hack to we don't always have to call fs.Embed
		if strings.HasPrefix(serveOn, "/pages/") {
			serveOn = strings.TrimPrefix(serveOn, "/pages")
		}
		if strings.HasSuffix(serveOn, ".html") {
			serveOn = strings.TrimSuffix(serveOn, ".html")
		}
		if strings.HasSuffix(serveOn, "/index") {
			serveOn = strings.TrimSuffix(serveOn, "index")
		}
		templateTokens := strings.Split(strings.Trim(serveOn, "/"), "/")

		if strings.HasSuffix(serveOn, "/_name") {
			serveOn = strings.TrimSuffix(serveOn, "_name")
		}
		templateMap := make(map[int]string)
		for i, s := range templateTokens {
			if strings.HasPrefix(s, "_") {
				templateMap[i] = strings.TrimPrefix(s, "_")
			}
		}
		servePage := func(ctx context.Context, req *components.Request) (components.Response, error) {
			tokens := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
			for i, s := range templateMap {
				req.PathParameters[s] = tokens[i]
			}
			klog.Infof("req %+v", req.URL.Path)
			klog.Infof("pathparameter %+v", req.PathParameters)
			return endpoint.ServeHTTP(ctx, req)
		}
		klog.Infof("serving %s on %s", p, serveOn)
		mux.HandleFunc(serveOn, s.ServeHTTP(servePage))
	}

	if info.IsDir() {
		if err := c.addHandlersFromDir(s, mux, p); err != nil {
			return err
		}
	}
	return nil
}

type TemplateEndpoint struct {
	server   *components.Server
	template templates.Template
}

func (e *TemplateEndpoint) ServeHTTP(ctx context.Context, req *components.Request) (components.Response, error) {
	data := e.server.NewScope(ctx)

	var b bytes.Buffer
	if err := e.template.RenderHTML(ctx, &b, req, data); err != nil {
		return nil, err
	}
	response := components.SimpleResponse{
		Body: b.Bytes(),
	}
	return response, nil
}
