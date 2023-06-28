package sanitizer

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"net/url"
	"regexp"
	"strings"

	cssparser "github.com/rollick/douceur/parser"
	"github.com/slt/douceur/css"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

type Client struct {
	ShowImages  bool
	Attachments map[string]string
}

var cssURLRegexp = regexp.MustCompile(`url\([^)]*\)`)

var allowedStyles = map[string]bool{
	"direction":       true,
	"font":            true,
	"font-family":     true,
	"font-style":      true,
	"font-variant":    true,
	"font-size":       true,
	"font-weight":     true,
	"letter-spacing":  true,
	"line-height":     true,
	"text-align":      true,
	"text-decoration": true,
	"text-indent":     true,
	"text-overflow":   true,
	"text-shadow":     true,
	"text-transform":  true,
	"white-space":     true,
	"word-spacing":    true,
	"word-wrap":       true,
	"vertical-align":  true,

	"color":             true,
	"background":        true,
	"background-color":  true,
	"background-image":  true,
	"background-repeat": true,

	"border":        true,
	"border-color":  true,
	"border-radius": true,
	"border-top":    true,
	"border-bottom": true,
	"border-right":  true,
	"border-left":   true,

	"height":         true,
	"margin":         true,
	"margin-right":   true,
	"margin-left":    true,
	"margin-bottom":  true,
	"margin-top":     true,
	"padding":        true,
	"padding-right":  true,
	"padding-left":   true,
	"padding-bottom": true,
	"padding-top":    true,
	"width":          true,
	"max-width":      true,
	"min-width":      true,
	"display":        true,

	"clear": true,
	"float": true,

	"border-collapse": true,
	"border-spacing":  true,
	"caption-side":    true,
	"empty-cells":     true,
	"table-layout":    true,

	"list-style-type":     true,
	"list-style-position": true,
}

func (sc *Client) SanitizeCSSRule(rule *css.Rule) {
	// Disallow @import
	if rule.Kind == css.AtRule && strings.EqualFold(rule.Name, "@import") {
		rule.Prelude = "url(about:blank)"
	}

	rule.Declarations = sc.SanitizeCSSDecls(rule.Declarations)

	for _, child := range rule.Rules {
		sc.SanitizeCSSRule(child)
	}
}

func (sc *Client) SanitizeCSSDecls(decls []*css.Declaration) []*css.Declaration {
	sanitized := make([]*css.Declaration, 0, len(decls))
	for _, decl := range decls {
		if !allowedStyles[decl.Property] {
			continue
		}

		decl.Value = cssURLRegexp.ReplaceAllString(decl.Value, "url(about:blank)")

		sanitized = append(sanitized, decl)
	}
	return sanitized
}

func (sc *Client) SanitizeImageURL(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return "about:blank"
	}

	switch strings.ToLower(u.Scheme) {
	case "cid":
		if u.Opaque == "" || !sc.ShowImages {
			return "about:blank"
		}

		table := crc32.MakeTable(crc32.IEEE)
		checksum := crc32.Checksum([]byte(u.Opaque), table)
		checksumSt := fmt.Sprintf("%x", checksum)

		if v, ok := sc.Attachments[checksumSt]; ok {
			return "/api/v1/attachments/" + v + "/cid"
		} else {
			return "about:blank"
		}
	default:
		if sc.ShowImages {
			return src
		} else {
			return "about:blank"
		}
	}
}

func (sc *Client) SanitizeNode(n *html.Node) {
	if n.Type == html.ElementNode {
		if strings.EqualFold(n.Data, "img") {
			for i := range n.Attr {
				attr := &n.Attr[i]
				if strings.EqualFold(attr.Key, "src") {
					attr.Val = sc.SanitizeImageURL(attr.Val)
				}
			}
		} else if strings.EqualFold(n.Data, "style") {
			var s string
			c := n.FirstChild
			for c != nil {
				if c.Type == html.TextNode {
					s += c.Data
				}

				next := c.NextSibling
				n.RemoveChild(c)
				c = next
			}

			stylesheet, err := cssparser.Parse(s)
			if err != nil {
				s = ""
			} else {
				for _, rule := range stylesheet.Rules {
					sc.SanitizeCSSRule(rule)
				}

				s = stylesheet.String()
			}

			n.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: s,
			})
		}

		for i := range n.Attr {
			attr := &n.Attr[i]

			if strings.EqualFold(attr.Key, "style") {
				decls, err := cssparser.ParseDeclarations(attr.Val)

				if err != nil {
					attr.Val = ""
					continue
				}

				decls = sc.SanitizeCSSDecls(decls)

				attr.Val = ""
				for _, d := range decls {
					attr.Val += d.String()
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sc.SanitizeNode(c)
	}
}

func (sc *Client) SanitizeHTML(b []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	sc.SanitizeNode(doc)

	var prerender bytes.Buffer
	if err := html.Render(&prerender, doc); err != nil {
		return nil, fmt.Errorf("failed to render HTML: %v", err)
	}

	/*
		input, err := inliner.Inline(prerender.String())
		if err != nil {
			return nil, fmt.Errorf("failed to prepare HTML: %v", err)
		}
	*/

	p := bluemonday.UGCPolicy()

	p.AllowElements("style")
	p.AllowAttrs("style").Globally()
	p.AllowAttrs("align").Globally()

	p.AllowAttrs("cellspacing").OnElements("table")
	p.AllowAttrs("cellpadding").OnElements("table")

	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.RequireNoFollowOnLinks(true)

	return p.SanitizeBytes(prerender.Bytes()), nil
}
