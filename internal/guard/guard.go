// Package guard decides which paths noether may open.
package guard

import (
	"path/filepath"
	"regexp"
	"strings"
)

var secretPatterns = []string{
	"**/.env*",
	"**/*.pem",
	"**/*.key",
	"**/*.p12",
	"**/id_rsa*",
	"**/id_ed25519*",
}

type Guard struct {
	denied []*regexp.Regexp
}

// New always includes common secret files, so a missing deny list never exposes them.
func New(deny []string) (Guard, error) {
	var g Guard
	for _, pattern := range append(secretPatterns, deny...) {
		re, err := compile(pattern)
		if err != nil {
			return Guard{}, err
		}
		g.denied = append(g.denied, re)
	}
	return g, nil
}

func (g Guard) Allows(path string) bool {
	slashed := filepath.ToSlash(path)
	for _, re := range g.denied {
		if re.MatchString(slashed) {
			return false
		}
	}
	return true
}

// compile turns a glob into a regular expression: ** crosses directories, * and ? do not.
// A pattern matches a path or anything beneath it, so "**/private" also denies its contents.
func compile(glob string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; {
		case strings.HasPrefix(glob[i:], "**/"):
			b.WriteString("(?:.*/)?")
			i += 2
		case strings.HasPrefix(glob[i:], "**"):
			b.WriteString(".*")
			i++
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("(?:/.*)?$")
	return regexp.Compile(b.String())
}
