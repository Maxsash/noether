// Package server shows noether's reports as a web page on this machine.
package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/maxsash/noether/internal/config"
	"github.com/maxsash/noether/internal/discover"
	"github.com/maxsash/noether/internal/guard"
	"github.com/maxsash/noether/internal/record"
	"github.com/maxsash/noether/internal/repo"
)

//go:embed assets
var assets embed.FS

type Server struct {
	config config.Config
	guard  guard.Guard
	editor editor
	now    func() time.Time

	mu      sync.Mutex
	sources map[string]repo.Source
}

func New(c config.Config, g guard.Guard) *Server {
	e := editor{template: c.Editor}
	if c.Editor == config.NoEditor {
		e.template = ""
	}
	return &Server{config: c, guard: g, editor: e, now: time.Now, sources: map[string]repo.Source{}}
}

// Handler answers only requests addressed to this machine on port, so a web page elsewhere
// cannot reach it by pointing a domain name at 127.0.0.1.
func (s *Server) Handler(port string) http.Handler {
	static, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(static))
	mux.HandleFunc("GET /api/repos", s.list)
	mux.HandleFunc("GET /api/repos/{id...}", s.detail)
	return localOnly(port, mux)
}

func localOnly(port string, next http.Handler) http.Handler {
	hosts := map[string]bool{"127.0.0.1:" + port: true, "localhost:" + port: true, "[::1]:" + port: true}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hosts[r.Host] {
			http.Error(w, "noether only answers requests to this machine", http.StatusForbidden)
			return
		}
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type listView struct {
	Repos    []repo.Summary `json:"repos"`
	Problems []string       `json:"problems"`
}

func (s *Server) list(w http.ResponseWriter, _ *http.Request) {
	sources, problems := s.discover()
	reports := repo.BuildAll(sources, s.guard, s.now())
	view := listView{Repos: make([]repo.Summary, len(reports)), Problems: append([]string{}, problems...)}
	for i, r := range reports {
		view.Repos[i] = r.Summary()
	}
	writeJSON(w, view)
}

func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	src, ok := s.source(r.PathValue("id"))
	if !ok {
		s.discover()
		if src, ok = s.source(r.PathValue("id")); !ok {
			http.NotFound(w, r)
			return
		}
	}
	writeJSON(w, s.detailView(repo.Build(src, s.guard, s.now())))
}

func (s *Server) discover() ([]repo.Source, []string) {
	sources, problems := discover.Sources(s.config, s.guard)
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.sources)
	for _, src := range sources {
		s.sources[src.ID] = src
	}
	return sources, problems
}

func (s *Server) source(id string) (repo.Source, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	src, ok := s.sources[id]
	return src, ok
}

type detailView struct {
	repo.Report
	StateHTML string     `json:"stateHtml,omitempty"`
	StateOpen string     `json:"stateOpen,omitempty"`
	Tasks     []taskView `json:"tasks"`
	Docs      []docView  `json:"docs"`
}

type taskView struct {
	record.Task
	HTML string `json:"html"`
	Open string `json:"open,omitempty"`
}

type docView struct {
	repo.Doc
	Open string `json:"open,omitempty"`
}

func (s *Server) detailView(r repo.Report) detailView {
	abs := func(rel string) string { return filepath.Join(r.Path, filepath.FromSlash(rel)) }
	view := detailView{Report: r, Tasks: []taskView{}, Docs: []docView{}}
	if r.State != nil {
		view.StateHTML = s.editor.render(r.StateText, abs(r.State.Path))
		view.StateOpen = s.editor.link(abs(r.State.Path), 1)
	}
	for _, t := range r.Tasks {
		view.Tasks = append(view.Tasks, taskView{Task: t, HTML: s.editor.inline(t.Text, abs(t.File)), Open: s.editor.link(abs(t.File), t.Line)})
	}
	for _, d := range r.Docs {
		view.Docs = append(view.Docs, docView{Doc: d, Open: s.editor.link(abs(d.Path), 1)})
	}
	return view
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
