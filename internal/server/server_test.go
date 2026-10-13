package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maxsash/noether/internal/config"
	"github.com/maxsash/noether/internal/guard"
)

func newTestServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "demo")
	if out, err := exec.Command("git", "init", "-q", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	files := map[string]string{
		"docs/handoff.md": "# Handoff\n\nSee [testing](testing.md) and <script>alert(1)</script> [bad](javascript:alert(1)).\n",
		"TODO.md":         "- [ ] write the **guide**\n",
	}
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c, err := config.ForFolder(root)
	if err != nil {
		t.Fatal(err)
	}
	g, err := guard.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return New(c, g).Handler("4747"), dir
}

func get(t *testing.T, h http.Handler, host, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Host = host
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRefusesOtherHosts(t *testing.T) {
	h, _ := newTestServer(t)
	for _, host := range []string{"evil.example:4747", "127.0.0.1:9999", "localhost"} {
		if rec := get(t, h, host, "/api/repos"); rec.Code != http.StatusForbidden {
			t.Errorf("host %s: status %d, want 403", host, rec.Code)
		}
	}
	if rec := get(t, h, "localhost:4747", "/"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Noether") {
		t.Errorf("page: status %d", rec.Code)
	}
}

func TestAPI(t *testing.T) {
	h, dir := newTestServer(t)
	rec := get(t, h, "127.0.0.1:4747", "/api/repos")
	var list listView
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("list: %v\n%s", err, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"problems":[]`) {
		t.Errorf("problems should be an empty list, not null: %s", rec.Body)
	}
	if len(list.Repos) != 1 || list.Repos[0].ID != "demo" || list.Repos[0].OpenTasks != 1 {
		t.Fatalf("list = %+v", list)
	}

	rec = get(t, h, "127.0.0.1:4747", "/api/repos/demo")
	var detail struct {
		StateHTML string `json:"stateHtml"`
		Tasks     []struct {
			HTML string `json:"html"`
			Open string `json:"open"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("detail: %v\n%s", err, rec.Body)
	}
	if strings.Contains(detail.StateHTML, "<script>") || strings.Contains(detail.StateHTML, "javascript:") {
		t.Errorf("state HTML kept something dangerous:\n%s", detail.StateHTML)
	}
	wantLink := `href="vscode://file` + filepath.ToSlash(filepath.Join(dir, "docs", "testing.md")) + `:1"`
	if !strings.Contains(detail.StateHTML, wantLink) {
		t.Errorf("relative link not sent to the editor; want %s in\n%s", wantLink, detail.StateHTML)
	}
	if len(detail.Tasks) != 1 || detail.Tasks[0].HTML != "write the <strong>guide</strong>" || !strings.HasSuffix(detail.Tasks[0].Open, "TODO.md:1") {
		t.Errorf("tasks = %+v", detail.Tasks)
	}

	if rec := get(t, h, "127.0.0.1:4747", "/api/repos/unknown"); rec.Code != http.StatusNotFound {
		t.Errorf("unknown repository: status %d", rec.Code)
	}
}

func TestEditorLink(t *testing.T) {
	e := editor{template: config.DefaultEditor}
	if got, want := e.link("/Users/me/ship shooter/README.md", 0), "vscode://file/Users/me/ship%20shooter/README.md:1"; got != want {
		t.Errorf("link = %s, want %s", got, want)
	}
	if got := (editor{}).link("/a.md", 3); got != "" {
		t.Errorf("links turned off, got %s", got)
	}
}
