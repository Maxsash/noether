package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	c, err := Parse(`
roots = ["~/dev"]
deny = ["~/dev/secret/**", "**/notes"]

[defaults]
state_lines = 120

[defaults.upkeep]
review = { every = "30d", commits = 40 }

[[repo]]
path = "~/elsewhere/app"
tasks = ["Docs/WorkPlan.md"]

[repo.upkeep]
review = { every = "14d" }
`)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "dev"); c.Roots[0] != want {
		t.Errorf("root = %s, want %s", c.Roots[0], want)
	}
	if want := filepath.Join(home, "dev/secret/**"); c.Deny[0] != want || c.Deny[1] != "**/notes" {
		t.Errorf("deny = %v", c.Deny)
	}
	if c.Depth != defaultDepth || c.Editor != DefaultEditor {
		t.Errorf("defaults not applied: depth %d, editor %q", c.Depth, c.Editor)
	}
	app := filepath.Join(home, "elsewhere/app")
	settings := c.SettingsFor(app)
	if settings.Tasks[0] != "Docs/WorkPlan.md" || settings.StateLines != 120 || settings.Upkeep["review"].Every != "14d" {
		t.Errorf("settings for app = %+v", settings)
	}
	if other := c.SettingsFor("/somewhere/else"); other.Upkeep["review"].Commits != 40 || len(other.Tasks) != 0 {
		t.Errorf("settings for another repo = %+v", other)
	}
}

func TestParseRejectsMistakes(t *testing.T) {
	for name, text := range map[string]string{
		"unknown key":    "root = [\"~/dev\"]\n",
		"bad cadence":    "[defaults.upkeep]\nreview = { every = \"monthly\" }\n",
		"bad repo entry": "[[repo]]\npath = \"/x\"\n[repo.upkeep]\nreview = { every = \"0d\" }\n",
		"not toml":       "roots = [\n",
	} {
		if _, err := Parse(text); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "absent.toml"))
	if !IsMissing(err) {
		t.Errorf("Load of a missing file = %v, want a missing-file error", err)
	}
}
