// Package discover finds the repositories noether reports on.
package discover

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/maxsash/noether/internal/config"
	"github.com/maxsash/noether/internal/guard"
	"github.com/maxsash/noether/internal/repo"
)

// heavy folders hold dependencies or build output, never a repository worth listing.
var heavy = map[string]bool{
	"node_modules": true, "vendor": true, "_vendor": true, "_to_delete": true,
	"build": true, "dist": true, "target": true, "coverage": true,
	"DerivedData": true, "Pods": true, "test-results": true,
}

// Sources lists every repository under the roots plus any the config names directly.
// Problems are roots or named repositories that could not be read.
func Sources(c config.Config, g guard.Guard) (sources []repo.Source, problems []string) {
	var paths []string
	for _, root := range c.Roots {
		found, err := repositories(root, g, c.Depth)
		if err != nil {
			problems = append(problems, err.Error())
		}
		paths = append(paths, found...)
	}
	for _, r := range c.Repos {
		if isRepository(r.Path) {
			paths = append(paths, r.Path)
		} else {
			problems = append(problems, fmt.Sprintf("%s is not a git repository", r.Path))
		}
	}
	slices.Sort(paths)
	paths = slices.Compact(paths)
	ids := Names(paths)
	for _, p := range paths {
		sources = append(sources, repo.Source{ID: ids[p], Path: p, Settings: c.SettingsFor(p)})
	}
	return sources, problems
}

// repositories walks at most depth folders below root. Hidden and heavy folders are skipped;
// nested repositories are still found.
func repositories(root string, g guard.Guard, depth int) ([]string, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}
	var found []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" {
			found = append(found, filepath.Dir(p))
			return filepath.SkipDir
		}
		if p == root {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") || heavy[d.Name()] || !g.Allows(p) || levels(root, p) > depth {
			return filepath.SkipDir
		}
		return nil
	})
	return found, err
}

func isRepository(p string) bool {
	info, err := os.Stat(filepath.Join(p, ".git"))
	return err == nil && info.IsDir()
}

func levels(root, p string) int {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return 0
	}
	return strings.Count(filepath.ToSlash(rel), "/") + 1
}

// Names gives each path the shortest run of trailing folders that no other path shares.
func Names(paths []string) map[string]string {
	names := map[string]string{}
	for _, p := range paths {
		parts := strings.Split(filepath.ToSlash(p), "/")
		for n := 1; n <= len(parts); n++ {
			candidate := strings.Join(parts[len(parts)-n:], "/")
			if !sharedSuffix(candidate, p, paths) {
				names[p] = candidate
				break
			}
		}
	}
	return names
}

func sharedSuffix(suffix, self string, paths []string) bool {
	for _, other := range paths {
		if other != self && (filepath.ToSlash(other) == suffix || strings.HasSuffix(filepath.ToSlash(other), "/"+suffix)) {
			return true
		}
	}
	return false
}
