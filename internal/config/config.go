// Package config reads the owner's noether config: where to look for repositories, what
// never to open, and the record settings every repository starts from.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/maxsash/noether/internal/record"
)

type Config struct {
	Roots    []string        `toml:"roots"`
	Depth    int             `toml:"depth"`
	Deny     []string        `toml:"deny"`
	Editor   string          `toml:"editor"`
	Defaults record.Settings `toml:"defaults"`
	Repos    []Repo          `toml:"repo"`
}

// Repo adds a repository outside the roots, or gives one settings it does not keep itself.
type Repo struct {
	Path string `toml:"path"`
	record.Settings
}

const (
	DefaultEditor = "vscode://file{path}:{line}"
	NoEditor      = "none"
	defaultDepth  = 4
)

// DefaultPath follows the XDG convention on every platform, so the file is easy to find.
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "noether", "config.toml"), nil
}

// Load returns an error wrapping fs.ErrNotExist when there is no file at path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	c, err := Parse(string(data))
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	return c, nil
}

func Parse(text string) (Config, error) {
	var c Config
	meta, err := toml.Decode(text, &c)
	if err != nil {
		return Config{}, err
	}
	if unknown := meta.Undecoded(); len(unknown) > 0 {
		return Config{}, fmt.Errorf("unknown keys %v", unknown)
	}
	return c.normalized()
}

// ForFolder is the config used when there is no config file: just the given folder.
func ForFolder(dir string) (Config, error) {
	return Config{Roots: []string{dir}}.normalized()
}

func IsMissing(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}

// SettingsFor gives a repository the defaults, overlaid with any [[repo]] entry for it.
func (c Config) SettingsFor(path string) record.Settings {
	settings := c.Defaults
	for _, r := range c.Repos {
		if r.Path == path {
			settings = settings.Merge(r.Settings)
		}
	}
	return settings
}

func (c Config) normalized() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	for i, root := range c.Roots {
		if c.Roots[i], err = absolute(root, home); err != nil {
			return Config{}, err
		}
	}
	for i, pattern := range c.Deny {
		c.Deny[i] = expandHome(pattern, home)
	}
	for i := range c.Repos {
		if c.Repos[i].Path, err = absolute(c.Repos[i].Path, home); err != nil {
			return Config{}, err
		}
		if err := c.Repos[i].Validate(); err != nil {
			return Config{}, fmt.Errorf("repo %s: %w", c.Repos[i].Path, err)
		}
	}
	if err := c.Defaults.Validate(); err != nil {
		return Config{}, fmt.Errorf("defaults: %w", err)
	}
	if c.Depth <= 0 {
		c.Depth = defaultDepth
	}
	if c.Editor == "" {
		c.Editor = DefaultEditor
	}
	return c, nil
}

func absolute(p, home string) (string, error) {
	return filepath.Abs(expandHome(p, home))
}

func expandHome(p, home string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[1:])
	}
	return p
}
