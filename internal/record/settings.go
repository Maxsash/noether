// Package record reads the plain-text record a repository keeps about itself: where it
// stands, what is left to do, and when upkeep jobs last ran.
package record

import (
	"fmt"
	"maps"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/maxsash/noether/internal/upkeep"
)

// Settings names a repository's record files and its upkeep schedule.
type Settings struct {
	State      string                    `toml:"state"`
	Tasks      []string                  `toml:"tasks"`
	Log        string                    `toml:"log"`
	StateLines int                       `toml:"state_lines"`
	Upkeep     map[string]upkeep.Cadence `toml:"upkeep"`
}

// ParseSettings reads the TOML inside a repository's noether block.
func ParseSettings(text string) (Settings, error) {
	var s Settings
	meta, err := toml.Decode(text, &s)
	if err != nil {
		return Settings{}, err
	}
	if unknown := meta.Undecoded(); len(unknown) > 0 {
		return Settings{}, fmt.Errorf("unknown keys %v", unknown)
	}
	return s, s.Validate()
}

func (s Settings) Validate() error {
	for _, kind := range s.Kinds() {
		if _, err := s.Upkeep[kind].Days(); err != nil {
			return fmt.Errorf("upkeep %s: %w", kind, err)
		}
	}
	return nil
}

// Merge returns s with every field that o sets taking precedence. Upkeep jobs merge by name.
func (s Settings) Merge(o Settings) Settings {
	if o.State != "" {
		s.State = o.State
	}
	if len(o.Tasks) > 0 {
		s.Tasks = o.Tasks
	}
	if o.Log != "" {
		s.Log = o.Log
	}
	if o.StateLines > 0 {
		s.StateLines = o.StateLines
	}
	if len(o.Upkeep) > 0 {
		merged := maps.Clone(s.Upkeep)
		if merged == nil {
			merged = map[string]upkeep.Cadence{}
		}
		maps.Copy(merged, o.Upkeep)
		s.Upkeep = merged
	}
	return s
}

func (s Settings) Kinds() []string {
	kinds := make([]string, 0, len(s.Upkeep))
	for kind := range s.Upkeep {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

var (
	stateNames = []string{"handoff", "status"}
	taskNames  = []string{"todo", "followups", "workplan", "tasks", "backlog"}
	logNames   = []string{"upkeep", "log"}
)

// Detect names record files by convention, for repositories that do not name their own.
func Detect(markdownFiles []string) Settings {
	byStem := map[string][]string{}
	for _, f := range markdownFiles {
		byStem[stem(f)] = append(byStem[stem(f)], f)
	}
	s := Settings{State: first(byStem, stateNames), Log: first(byStem, logNames)}
	for _, name := range taskNames {
		s.Tasks = append(s.Tasks, byStem[name]...)
	}
	return s
}

func first(byStem map[string][]string, names []string) string {
	for _, name := range names {
		if files := byStem[name]; len(files) > 0 {
			return files[0]
		}
	}
	return ""
}

func stem(file string) string {
	base := strings.ToLower(strings.TrimSuffix(path.Base(file), path.Ext(file)))
	return strings.NewReplacer("-", "", "_", "", " ", "").Replace(base)
}
