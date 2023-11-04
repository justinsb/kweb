package templates

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/components/users"
	"github.com/justinsb/kweb/templates/scopes"
	"golang.org/x/net/html"
)

type Template struct {
	Data []byte
}

func (s *Template) RenderHTML(ctx context.Context, w io.Writer, server *components.Server, req *components.Request) error {
	page, err := html.Parse(bytes.NewReader([]byte(defaultPage)))
	if err != nil {
		return fmt.Errorf("failed to parse page: %w", err)
	}

	slot := findSlot(page, "body")
	if slot == nil {
		return fmt.Errorf("failed to find slot %q: %w", "body", err)
	}

	nodes, err := html.ParseFragment(bytes.NewReader(s.Data), slot)
	if err != nil {
		return fmt.Errorf("failed to parse html: %w", err)
	}
	for _, node := range nodes {
		slot.Parent.InsertBefore(node, slot)
	}
	slot.Parent.RemoveChild(slot)

	data := scopes.NewScope()

	// TODO: Move to component
	data.Values["user"] = scopes.Value{
		Function: func() interface{} {
			return users.GetUser(ctx)
		},
	}

	data.Values["components"] = scopes.Value{
		Function: func() interface{} {
			m := make(map[string]any)
			for _, component := range server.Components {
				v := component.ScopeValues()
				if v != nil {
					key := component.Key()
					m[key] = v
				}
			}
			return m
		},
	}

	var render Render
	bw := bufio.NewWriter(w)
	render.w = bw
	render.data = data

	if err := render.renderNode(page); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	return nil
}

const defaultPage = `
<!doctype html>
<html>
<head>
	<meta charset="utf-8">
	<title>My Example App</title>
</head>
<body>
	<slot name="body"></slot>
</body>
</html>
`

func findSlot(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && node.Data == "slot" {
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == name {
				return node
			}
		}
	}

	child := node.FirstChild
	for child != nil {
		found := findSlot(child, name)
		if found != nil {
			return found
		}
		child = child.NextSibling
	}

	return nil
}
