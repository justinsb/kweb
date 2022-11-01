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
	"k8s.io/klog/v2"
)

type Options struct {
	Base fs.FS
}

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

		endpoint := &TemplateEndpoint{template: template}
		serveOn := "/" + p
		// Hack to we don't always have to call fs.Embed
		if strings.HasPrefix(serveOn, "/pages/") {
			serveOn = strings.TrimPrefix(serveOn, "/pages")
		}
		if strings.HasSuffix(serveOn, "/index.html") {
			serveOn = strings.TrimSuffix(serveOn, "index.html")
		}

		klog.Infof("serving %s on %s", p, serveOn)
		mux.HandleFunc(serveOn, s.ServeHTTP(endpoint.ServeHTTP))
	}

	if info.IsDir() {
		if err := c.addHandlersFromDir(s, mux, p); err != nil {
			return nil
		}
	}
	return nil
}

type TemplateEndpoint struct {
	template templates.Template
}

func (e *TemplateEndpoint) ServeHTTP(ctx context.Context, req *components.Request) (components.Response, error) {
	var b bytes.Buffer
	if err := e.template.RenderHTML(ctx, &b, req); err != nil {
		return nil, err
	}
	response := components.SimpleResponse{
		Body: b.Bytes(),
	}
	return response, nil

}