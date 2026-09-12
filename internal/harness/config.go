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
	ReaperC2URL    string
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
		"Config{ReaperURL:%q ReaperC2URL:%q ReaperUsername:%q ReaperPassword:%s Client:%q Engagement:%q SessionsDir:%q Objectives:%v DryRun:%v}",
		c.ReaperURL, c.ReaperC2URL, c.ReaperUsername, passwordState, c.Client, c.Engagement, c.SessionsDir, c.Objectives, c.DryRun,
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
const UsageText = `Usage: harness --reaper-url URL --reaper-c2-url URL --reaper-username USER \
         --engagement NAME --objective "text" [--objective "text" ...] [options]

Required:
  --reaper-url URL          ReaperC2 admin panel / operator dashboard base URL
                            (e.g. https://c2.example.com:8443)
  --reaper-c2-url URL       Beacon listener / implant C2 base URL
                            (e.g. https://c2.example.com:8080). Must differ from
                            --reaper-url; beacons phone home here, not the dashboard.
  --reaper-username USER    ReaperC2 operator username
  --engagement NAME         Engagement / workspace name this run is scoped to
  --objective TEXT          One engagement objective. Repeat for multiple. At least one required.

One of these is required for the password (never pass it as a bare CLI arg in shared shells):
  --reaper-password PASS    ReaperC2 operator password
  REAPER_PASSWORD env var   Same, via environment
  (omit both to be prompted interactively, input hidden — requires a TTY)

Optional:
  --client NAME             Client / customer name for the report
  --objectives-file PATH    Read additional objectives, one per line
  --sessions-dir PATH       Where to write the session prompt file (default: sessions)
  --dry-run                 Build and print the prompt, don't launch claude
  -h, --help                Show this help

Examples:
  harness --reaper-url https://c2.internal:8443 --reaper-c2-url https://c2.internal:8080 \
    --reaper-username op1 --client "Acme Corp" --engagement "acme-2026-q3" \
    --objective "Obtain domain admin from an external foothold" \
    --objective "Demonstrate access to the finance file share"
`

// ParseArgs parses CLI flags, falling back to REAPER_URL / REAPER_C2_URL /
// REAPER_USERNAME / REAPER_PASSWORD / REAPER_ENGAGEMENT environment variables
// for the connection and scope fields. It does not validate or prompt for a
// password — see Validate and ResolvePassword.
func ParseArgs(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("harness", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, UsageText) }

	cfg := &Config{
		ReaperURL:      os.Getenv("REAPER_URL"),
		ReaperC2URL:    os.Getenv("REAPER_C2_URL"),
		ReaperUsername: os.Getenv("REAPER_USERNAME"),
		ReaperPassword: os.Getenv("REAPER_PASSWORD"),
		Engagement:     os.Getenv("REAPER_ENGAGEMENT"),
	}

	var objectives stringSlice
	var objectivesFile string

	fs.StringVar(&cfg.ReaperURL, "reaper-url", cfg.ReaperURL, "ReaperC2 admin panel base URL")
	fs.StringVar(&cfg.ReaperC2URL, "reaper-c2-url", cfg.ReaperC2URL, "ReaperC2 beacon listener / implant C2 base URL")
	fs.StringVar(&cfg.ReaperUsername, "reaper-username", cfg.ReaperUsername, "ReaperC2 operator username")
	fs.StringVar(&cfg.ReaperPassword, "reaper-password", cfg.ReaperPassword, "ReaperC2 operator password")
	fs.StringVar(&cfg.Client, "client", "", "Client / customer name")
	fs.StringVar(&cfg.Engagement, "engagement", cfg.Engagement, "Engagement name this run is scoped to")
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
	if c.ReaperC2URL == "" {
		missing = append(missing, "--reaper-c2-url")
	}
	if c.ReaperUsername == "" {
		missing = append(missing, "--reaper-username")
	}
	if c.Engagement == "" {
		missing = append(missing, "--engagement")
	}
	if len(c.Objectives) == 0 {
		missing = append(missing, "--objective (at least one)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required input(s): %s", strings.Join(missing, " "))
	}

	if err := validateHTTPURL("--reaper-url", c.ReaperURL); err != nil {
		return err
	}
	if err := validateHTTPURL("--reaper-c2-url", c.ReaperC2URL); err != nil {
		return err
	}
	if normalizeURL(c.ReaperURL) == normalizeURL(c.ReaperC2URL) {
		return fmt.Errorf("--reaper-c2-url is the beacon listener and must differ from the admin panel --reaper-url")
	}

	return nil
}

func validateHTTPURL(flagName, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("%s must start with http:// or https:// (got: %s)", flagName, raw)
	}
	return nil
}

func normalizeURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

// Finalize fills in defaults that depend on runtime state (a placeholder
// client name). Call after Validate succeeds. Engagement is required and is
// never generated here — the run stays scoped to the name the operator gave.
func (c *Config) Finalize() {
	if c.Client == "" {
		c.Client = "<UPDATE ME>"
	}
	if c.SessionsDir == "" {
		c.SessionsDir = "sessions"
	}
}
