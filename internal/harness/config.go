// Package harness builds the initial Claude Code session for one ai-harness
// engagement: it validates the required ReaperC2 connection details and
// objectives, resolves the operator password without ever writing it to disk,
// renders the session prompt, and launches claude with the right environment.
package harness

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// Config holds the inputs for one harness run.
type Config struct {
	ReaperURL      string
	ReaperUsername string
	ReaperPassword string
	Client         string
	Engagement     string
	SessionsDir    string
	Objectives     []string
	DryRun         bool
}

// String implements fmt.Stringer with ReaperPassword redacted, so that
// formatting a Config with %v/%+v — in a log line, an error, a test failure
// message, anywhere — can never leak the credential. Go's fmt package uses
// this instead of reflecting the struct's fields for both Config and *Config.
func (c Config) String() string {
	passwordState := "<empty>"
	if c.ReaperPassword != "" {
		passwordState = "<redacted>"
	}
	return fmt.Sprintf(
		"Config{ReaperURL:%q ReaperUsername:%q ReaperPassword:%s Client:%q Engagement:%q SessionsDir:%q Objectives:%v DryRun:%v}",
		c.ReaperURL, c.ReaperUsername, passwordState, c.Client, c.Engagement, c.SessionsDir, c.Objectives, c.DryRun,
	)
}

// stringSlice implements flag.Value to support a repeatable --objective flag.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// UsageText is the full CLI help text, shared by -h/--help and validation errors.
const UsageText = `Usage: harness --reaper-url URL --reaper-username USER \
         --objective "text" [--objective "text" ...] [options]

Required:
  --reaper-url URL          ReaperC2 admin panel base URL (e.g. https://c2.example.com:8443)
  --reaper-username USER    ReaperC2 operator username
  --objective TEXT          One engagement objective. Repeat for multiple. At least one required.

One of these is required for the password (never pass it as a bare CLI arg in shared shells):
  --reaper-password PASS    ReaperC2 operator password
  REAPER_PASSWORD env var   Same, via environment
  (omit both to be prompted interactively, input hidden — requires a TTY)

Optional:
  --client NAME             Client / customer name for the report
  --engagement NAME         Engagement name (default: eng-<timestamp>)
  --objectives-file PATH    Read additional objectives, one per line
  --sessions-dir PATH       Where to write the session prompt file (default: sessions)
  --dry-run                 Build and print the prompt, don't launch claude
  -h, --help                Show this help

Examples:
  harness --reaper-url https://c2.internal:8443 --reaper-username op1 \
    --client "Acme Corp" --engagement "acme-2026-q3" \
    --objective "Obtain domain admin from an external foothold" \
    --objective "Demonstrate access to the finance file share"
`

// ParseArgs parses CLI flags, falling back to REAPER_URL / REAPER_USERNAME /
// REAPER_PASSWORD environment variables for the connection fields. It does not
// validate or prompt for a password — see Validate and ResolvePassword.
func ParseArgs(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("harness", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, UsageText) }

	cfg := &Config{
		ReaperURL:      os.Getenv("REAPER_URL"),
		ReaperUsername: os.Getenv("REAPER_USERNAME"),
		ReaperPassword: os.Getenv("REAPER_PASSWORD"),
	}

	var objectives stringSlice
	var objectivesFile string

	fs.StringVar(&cfg.ReaperURL, "reaper-url", cfg.ReaperURL, "ReaperC2 admin panel base URL")
	fs.StringVar(&cfg.ReaperUsername, "reaper-username", cfg.ReaperUsername, "ReaperC2 operator username")
	fs.StringVar(&cfg.ReaperPassword, "reaper-password", cfg.ReaperPassword, "ReaperC2 operator password")
	fs.StringVar(&cfg.Client, "client", "", "Client / customer name")
	fs.StringVar(&cfg.Engagement, "engagement", "", "Engagement name")
	fs.StringVar(&cfg.SessionsDir, "sessions-dir", "sessions", "Where to write the session prompt file")
	fs.Var(&objectives, "objective", "Engagement objective (repeatable)")
	fs.StringVar(&objectivesFile, "objectives-file", "", "Path to a file of objectives, one per line")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Build and print the prompt, don't launch claude")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg.Objectives = append(cfg.Objectives, objectives...)

	if objectivesFile != "" {
		fromFile, err := readObjectivesFile(objectivesFile)
		if err != nil {
			return nil, fmt.Errorf("reading --objectives-file: %w", err)
		}
		cfg.Objectives = append(cfg.Objectives, fromFile...)
	}

	return cfg, nil
}

func readObjectivesFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out, nil
}

// Validate checks the required, non-secret fields. It intentionally does not
// touch the password so callers can surface validation errors (or serve --help)
// without forcing an interactive prompt first.
func (c *Config) Validate() error {
	var missing []string
	if c.ReaperURL == "" {
		missing = append(missing, "--reaper-url")
	}
	if c.ReaperUsername == "" {
		missing = append(missing, "--reaper-username")
	}
	if len(c.Objectives) == 0 {
		missing = append(missing, "--objective (at least one)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required input(s): %s", strings.Join(missing, " "))
	}

	u, err := url.Parse(c.ReaperURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("--reaper-url must start with http:// or https:// (got: %s)", c.ReaperURL)
	}

	return nil
}

// Finalize fills in defaults that depend on runtime state (a generated
// engagement name, a placeholder client name). Call after Validate succeeds.
func (c *Config) Finalize(timestamp string) {
	if c.Engagement == "" {
		c.Engagement = "eng-" + timestamp
	}
	if c.Client == "" {
		c.Client = "<UPDATE ME>"
	}
	if c.SessionsDir == "" {
		c.SessionsDir = "sessions"
	}
}
