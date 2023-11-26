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
	m := &pageMux{
		s:        s,
		mux:      mux,
		patterns: make(map[string]*patternMux),
	}
	if err := m.addHandlersFromDir(c.options.Base, "."); err != nil {
		return err
	}

	for pattern, handler := range m.patterns {
		mux.HandleFunc(pattern, s.ServeHTTP(handler.ServeHTTP))
	}
	return nil
}

func (c *Component) AddToScope(ctx context.Context, scope *scopes.Scope) {
	for _, scopeFunc := range c.options.ScopeValues {
		scopeFunc(ctx, scope)
	}
}

type pageMux struct {
	s   *components.Server
	mux *http.ServeMux

	patterns map[string]*patternMux
}

type patternMux struct {
	handlers []patternHandler
}

type patternHandler struct {
	match       []string
	templateMap map[int]string
	handler     func(ctx context.Context, req *components.Request) (components.Response, error)
}

func (m *patternMux) ServeHTTP(ctx context.Context, req *components.Request) (components.Response, error) {
	tokens := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	for _, h := range m.handlers {
		klog.Infof("check match %+v", h.match)
		if len(h.match) != len(tokens) {
			continue
		}
		isMatch := true
		for i, s := range h.match {
			if s == "" {
				continue
			}
			if s != tokens[i] {
				isMatch = false
				break
			}
		}
		if !isMatch {
			continue
		}
		for i, s := range h.templateMap {
			req.PathParameters[s] = tokens[i]
		}
		return h.handler(ctx, req)
	}
	klog.Warningf("no match found for tokens %+v", tokens)
	return components.ErrorResponse(http.StatusNotFound), nil
}

func (m *pageMux) addHandlersFromDir(base fs.FS, p string) error {
	entries, err := fs.ReadDir(base, p)
	if err != nil {
		return fmt.Errorf("error from ReadDir(%q): %w", p, err)
	}

	for _, entry := range entries {
		name := path.Join(p, entry.Name())
		if err := m.addHandlers(base, name, entry); err != nil {
			return err
		}
	}

	return nil
}

func (m *pageMux) addHandlers(base fs.FS, p string, info fs.DirEntry) error {
	if !info.IsDir() {
		templateData, err := loadRaw(base, p)
		if err != nil {
			return fmt.Errorf("error reading %q: %w", p, err)
		}

		template := templates.Template{
			Data: []byte(templateData),
		}

		endpoint := &TemplateEndpoint{template: template, server: m.s}
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

		match := make([]string, len(templateTokens))
		templateMap := make(map[int]string)
		for i, s := range templateTokens {
			if strings.HasPrefix(s, "_") {
				templateMap[i] = strings.TrimPrefix(s, "_")
			} else {
				match[i] = s
			}
		}
		pattern := ""
		for _, s := range templateTokens {
			if strings.HasPrefix(s, "_") {
				pattern += "/"
				break
			}
			pattern += "/" + s
		}

		klog.Infof("serving %s on %s", p, serveOn)
		pm, ok := m.patterns[pattern]
		if !ok {
			pm = &patternMux{}
			m.patterns[pattern] = pm
		}
		pm.handlers = append(pm.handlers, patternHandler{
			match:       match,
			templateMap: templateMap,
			handler:     endpoint.ServeHTTP,
		})
	}

	if info.IsDir() {
		if err := m.addHandlersFromDir(base, p); err != nil {
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
