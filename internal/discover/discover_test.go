package discover

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/maxsash/noether/internal/config"
	"github.com/maxsash/noether/internal/guard"
)

func TestSources(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		"work/web",
		"work/web/nested",
		"personal/web",
		"deep/a/b/c/d/too-deep",
		"app/node_modules/pkg",
		".hidden/repo",
		"private/diary",
	} {
		if err := os.MkdirAll(filepath.Join(root, dir, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(outside, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	c, err := config.Parse("roots = [\"" + root + "\"]\ndeny = [\"**/private\"]\n" +
		"[[repo]]\npath = \"" + outside + "\"\n[[repo]]\npath = \"" + filepath.Join(root, "not-a-repo") + "\"\n")
	if err != nil {
		t.Fatal(err)
	}
	g, err := guard.New(c.Deny)
	if err != nil {
		t.Fatal(err)
	}
	sources, problems := Sources(c, g)
	var got []string
	for _, s := range sources {
		rel, _ := filepath.Rel(root, s.Path)
		got = append(got, rel+" as "+s.ID)
	}
	outsideRel, _ := filepath.Rel(root, outside)
	want := []string{
		"personal/web as personal/web",
		"work/web as work/web",
		"work/web/nested as nested",
		outsideRel + " as " + filepath.Base(outside),
	}
	if !sameItems(got, want) {
		t.Errorf("sources =\n%v\nwant\n%v", got, want)
	}
	if len(problems) != 1 {
		t.Errorf("problems = %v, want one for the path that is not a repository", problems)
	}
}

func TestNames(t *testing.T) {
	got := Names([]string{"/a/maxsashlabs/web", "/a/personal/web", "/a/b/family-tree", "/x/web/web"})
	want := map[string]string{
		"/a/maxsashlabs/web": "maxsashlabs/web",
		"/a/personal/web":    "personal/web",
		"/a/b/family-tree":   "family-tree",
		"/x/web/web":         "web/web",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Names = %v, want %v", got, want)
	}
}

func sameItems(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}
