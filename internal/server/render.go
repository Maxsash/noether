package server

import (
	"bytes"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// markdown leaves raw HTML out of what it renders, so a repository's files cannot inject
// script into the page.
var markdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

type editor struct {
	template string
}

// link opens a file at a line in the owner's editor; it is empty when links are turned off.
// {path} always starts with a slash, including Windows paths such as /C:/repo/README.md.
func (e editor) link(file string, line int) string {
	if e.template == "" {
		return ""
	}
	slashed := filepath.ToSlash(file)
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}
	escaped := (&url.URL{Path: slashed}).EscapedPath()
	return strings.NewReplacer("{path}", escaped, "{line}", strconv.Itoa(max(line, 1))).Replace(e.template)
}

// render turns markdown into HTML. Relative links point at files beside the document, so
// they are sent to the editor instead of this server.
func (e editor) render(source, docPath string) string {
	src := []byte(source)
	doc := markdown.Parser().Parse(text.NewReader(src))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := n.(*ast.Link); ok && entering {
			if target := e.relativeTarget(string(link.Destination), docPath); target != "" {
				link.Destination = []byte(target)
			}
		}
		return ast.WalkContinue, nil
	})
	var out bytes.Buffer
	if err := markdown.Renderer().Render(&out, src, doc); err != nil {
		return ""
	}
	return out.String()
}

// inline renders a single line of markdown, such as a task, without a wrapping paragraph.
func (e editor) inline(source, docPath string) string {
	html := strings.TrimSpace(e.render(source, docPath))
	html = strings.TrimPrefix(html, "<p>")
	return strings.TrimSuffix(html, "</p>")
}

func (e editor) relativeTarget(destination, docPath string) string {
	u, err := url.Parse(destination)
	if err != nil || u.IsAbs() || u.Host != "" || u.Path == "" || strings.HasPrefix(u.Path, "/") {
		return ""
	}
	return e.link(filepath.Join(filepath.Dir(docPath), filepath.FromSlash(u.Path)), 1)
}
