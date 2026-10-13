// Command noether shows where each of your repositories stands: what changed, what is left
// to do, and which upkeep jobs are due.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"text/tabwriter"
	"time"

	"github.com/maxsash/noether/internal/config"
	"github.com/maxsash/noether/internal/discover"
	"github.com/maxsash/noether/internal/guard"
	"github.com/maxsash/noether/internal/repo"
	"github.com/maxsash/noether/internal/server"
)

const usage = `noether shows where each of your repositories stands.

Usage:
  noether serve  [--config file] [--addr 127.0.0.1:4747] [--open]
  noether status [--config file] [--json]

Without a config file, noether reads the repositories under the current folder.
The config lives at ~/.config/noether/config.toml unless --config says otherwise.
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "noether:", err)
		os.Exit(1)
	}
}

func run(args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(out, usage)
		return nil
	}
	switch args[0] {
	case "serve":
		return serve(args[1:], out, errOut)
	case "status":
		return status(args[1:], out, errOut)
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func serve(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := flags.String("config", "", "config file")
	addr := flags.String("addr", "127.0.0.1:4747", "address to listen on")
	open := flags.Bool("open", false, "open the page in the default browser")
	if err := flags.Parse(args); err != nil {
		return err
	}
	c, g, err := setup(*configPath, errOut)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	url := "http://127.0.0.1:" + port + "/"
	fmt.Fprintf(out, "noether is showing your repositories at %s\n", url)
	if *open {
		openBrowser(url)
	}
	s := &http.Server{Handler: server.New(c, g).Handler(port), ReadHeaderTimeout: 5 * time.Second}
	return s.Serve(listener)
}

func status(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	configPath := flags.String("config", "", "config file")
	asJSON := flags.Bool("json", false, "print JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	c, g, err := setup(*configPath, errOut)
	if err != nil {
		return err
	}
	sources, problems := discover.Sources(c, g)
	for _, p := range problems {
		fmt.Fprintln(errOut, "noether:", p)
	}
	reports := repo.BuildAll(sources, g, time.Now())
	summaries := make([]repo.Summary, len(reports))
	for i, r := range reports {
		summaries[i] = r.Summary()
	}
	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(summaries)
	}
	return printTable(out, summaries, time.Now())
}

func printTable(out io.Writer, summaries []repo.Summary, now time.Time) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "REPOSITORY\tBRANCH\tWARNINGS\tDUE\tOPEN TASKS\tLAST ACTIVITY")
	for _, s := range summaries {
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%s\n", s.ID, s.Branch, s.Warnings, s.Due, s.OpenTasks, since(s.LastActivity, now))
	}
	return w.Flush()
}

func since(t, now time.Time) string {
	if t.IsZero() {
		return "never"
	}
	switch d := now.Sub(t); {
	case d < time.Hour:
		return "within the hour"
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// setup loads the config, falling back to the current folder when the default file is absent.
func setup(path string, errOut io.Writer) (config.Config, guard.Guard, error) {
	c, err := loadConfig(path, errOut)
	if err != nil {
		return config.Config{}, guard.Guard{}, err
	}
	g, err := guard.New(c.Deny)
	return c, g, err
}

func loadConfig(path string, errOut io.Writer) (config.Config, error) {
	if path != "" {
		return config.Load(path)
	}
	path, err := config.DefaultPath()
	if err != nil {
		return config.Config{}, err
	}
	c, err := config.Load(path)
	if !config.IsMissing(err) {
		return c, err
	}
	dir, err := os.Getwd()
	if err != nil {
		return config.Config{}, err
	}
	fmt.Fprintf(errOut, "noether: no config at %s; reading repositories under %s\n", path, dir)
	return config.ForFolder(dir)
}

func openBrowser(url string) {
	command := map[string][]string{
		"darwin":  {"open"},
		"windows": {"rundll32", "url.dll,FileProtocolHandler"},
	}[runtime.GOOS]
	if command == nil {
		command = []string{"xdg-open"}
	}
	_ = exec.Command(command[0], append(command[1:], url)...).Start()
}
