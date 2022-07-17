package templates

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/justinsb/kweb/templates/mustache"
	"github.com/justinsb/kweb/templates/scopes"
	"golang.org/x/net/html"
	"k8s.io/klog/v2"
)

type Render struct {
	w    *bufio.Writer
	data *scopes.Scope
}

func (r *Render) renderTextNode(node *html.Node) error {
	text := node.Data

	if strings.Contains(text, "{{") {
		el, err := mustache.ParseExpressionList(text)
		if err != nil {
			return err
		}
		expanded, err := el.Eval(r.data)
		if err != nil {
			return err
		}
		klog.Infof("found mustache: %#v", el)
		text = expanded
	}

	return escape(r.w, text)

}

var directiveAttribute = map[string]bool{
	"*ngfor": true,
	"*ngif":  true,
}

func (r *Render) renderElementNode(node *html.Node) error {
	// TODO: Handle directive attributes

	return r.renderElementNodeInner(node)
}

// This logic is based on the logic in golang's html.Render

func (r *Render) renderNode(n *html.Node) error {
	w := r.w

	// Render non-element nodes; these are the easy cases.
	switch n.Type {
	case html.ErrorNode:
		return errors.New("html: cannot render an ErrorNode node")
	case html.TextNode:
		return r.renderTextNode(n)
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := r.renderNode(c); err != nil {
				return err
			}
		}
		return nil
	case html.ElementNode:
		// No-op.
		return r.renderElementNode(n)
	case html.CommentNode:
		if _, err := w.WriteString("<!--"); err != nil {
			return err
		}
		if _, err := w.WriteString(n.Data); err != nil {
			return err
		}
		if _, err := w.WriteString("-->"); err != nil {
			return err
		}
		return nil
	case html.DoctypeNode:
		if _, err := w.WriteString("<!DOCTYPE "); err != nil {
			return err
		}
		if _, err := w.WriteString(n.Data); err != nil {
			return err
		}
		if n.Attr != nil {
			var p, s string
			for _, a := range n.Attr {
				switch a.Key {
				case "public":
					p = a.Val
				case "system":
					s = a.Val
				}
			}
			if p != "" {
				if _, err := w.WriteString(" PUBLIC "); err != nil {
					return err
				}
				if err := writeQuoted(w, p); err != nil {
					return err
				}
				if s != "" {
					if err := w.WriteByte(' '); err != nil {
						return err
					}
					if err := writeQuoted(w, s); err != nil {
						return err
					}
				}
			} else if s != "" {
				if _, err := w.WriteString(" SYSTEM "); err != nil {
					return err
				}
				if err := writeQuoted(w, s); err != nil {
					return err
				}
			}
		}
		return w.WriteByte('>')
	case html.RawNode:
		_, err := w.WriteString(n.Data)
		return err
	default:
		return errors.New("html: unknown node type")
	}
}

func (r *Render) renderElementNodeInner(n *html.Node) error {
	w := r.w

	// Render the <xxx> opening tag.
	if err := w.WriteByte('<'); err != nil {
		return err
	}
	if _, err := w.WriteString(n.Data); err != nil {
		return err
	}
	for _, a := range n.Attr {
		if directiveAttribute[a.Key] {
			continue
		}
		if err := w.WriteByte(' '); err != nil {
			return err
		}
		if a.Namespace != "" {
			if _, err := w.WriteString(a.Namespace); err != nil {
				return err
			}
			if err := w.WriteByte(':'); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(a.Key); err != nil {
			return err
		}
		if _, err := w.WriteString(`="`); err != nil {
			return err
		}
		if err := escape(w, a.Val); err != nil {
			return err
		}
		if err := w.WriteByte('"'); err != nil {
			return err
		}
	}
	if voidElements[n.Data] {
		if n.FirstChild != nil {
			return fmt.Errorf("html: void element <%s> has child nodes", n.Data)
		}
		_, err := w.WriteString("/>")
		return err
	}
	if err := w.WriteByte('>'); err != nil {
		return err
	}

	// Add initial newline where there is danger of a newline beging ignored.
	if c := n.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
		switch n.Data {
		case "pre", "listing", "textarea":
			if err := w.WriteByte('\n'); err != nil {
				return err
			}
		}
	}

	// Render any child nodes.
	switch n.Data {
	case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				if _, err := w.WriteString(c.Data); err != nil {
					return err
				}
			} else {
				if err := r.renderNode(c); err != nil {
					return err
				}
			}
		}
		// if n.Data == "plaintext" {
		// 	// Don't render anything else. <plaintext> must be the
		// 	// last element in the file, with no closing tag.
		// 	return plaintextAbort
		// }
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := r.renderNode(c); err != nil {
				return err
			}
		}
	}

	// Render the </xxx> closing tag.
	if _, err := w.WriteString("</"); err != nil {
		return err
	}
	if _, err := w.WriteString(n.Data); err != nil {
		return err
	}
	return w.WriteByte('>')
}

// writeQuoted writes s to w surrounded by quotes. Normally it will use double
// quotes, but if s contains a double quote, it will use single quotes.
// It is used for writing the identifiers in a doctype declaration.
// In valid HTML, they can't contain both types of quotes.
func writeQuoted(w *bufio.Writer, s string) error {
	var q byte = '"'
	if strings.Contains(s, `"`) {
		q = '\''
	}
	if err := w.WriteByte(q); err != nil {
		return err
	}
	if _, err := w.WriteString(s); err != nil {
		return err
	}
	if err := w.WriteByte(q); err != nil {
		return err
	}
	return nil
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"keygen": true, // "keygen" has been removed from the spec, but are kept here for backwards compatibility.
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

const escapedChars = "&'<>\"\r"

func escape(w *bufio.Writer, s string) error {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		if _, err := w.WriteString(s[:i]); err != nil {
			return err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			// "&#34;" is shorter than "&quot;".
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		if _, err := w.WriteString(esc); err != nil {
			return err
		}
		i = strings.IndexAny(s, escapedChars)
	}
	_, err := w.WriteString(s)
	return err
}
