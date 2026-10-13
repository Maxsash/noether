package guard

import "testing"

func TestAllows(t *testing.T) {
	g, err := New([]string{"**/life/private", "**/whats-app/index.js", "/abs/only/*.txt"})
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{
		"/home/me/repo/README.md":                true,
		"/home/me/repo/.env":                     false,
		"/home/me/repo/.env.local":               false,
		"/home/me/repo/certs/server.pem":         false,
		"/home/me/life/private":                  false,
		"/home/me/life/private/notes.md":         false,
		"/home/me/life/privateer.md":             true,
		"/home/me/dev/whats-app/index.js":        false,
		"/home/me/dev/whats-app/index.json":      true,
		"/abs/only/a.txt":                        false,
		"/abs/only/deeper/a.txt":                 true,
		"/elsewhere/abs/only/a.txt":              true,
		"/home/me/repo/docs/environment.md":      true,
		"/home/me/repo/keys/deploy.keyboard.txt": true,
	}
	for path, want := range cases {
		if got := g.Allows(path); got != want {
			t.Errorf("Allows(%q) = %v, want %v", path, got, want)
		}
	}
}
