package harness

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ErrPasswordRequired is returned when no password was supplied and stdin isn't a
// terminal, so an interactive prompt isn't possible (e.g. under docker compose
// without -it, or in a non-interactive CI job).
var ErrPasswordRequired = errors.New("password not provided and stdin is not a terminal; pass --reaper-password or set REAPER_PASSWORD")

// ResolvePassword leaves cfg.ReaperPassword untouched if already set (from a flag
// or the REAPER_PASSWORD env var). Otherwise it prompts on stderr and reads from
// stdin with echo disabled. The password is never written anywhere but the
// in-memory Config.
func ResolvePassword(cfg *Config, stdin *os.File, stderr io.Writer) error {
	if cfg.ReaperPassword != "" {
		return nil
	}

	if !term.IsTerminal(int(stdin.Fd())) {
		return ErrPasswordRequired
	}

	fmt.Fprintf(stderr, "ReaperC2 password for %s: ", cfg.ReaperUsername)
	pw, err := term.ReadPassword(int(stdin.Fd()))
	fmt.Fprintln(stderr)
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}
	if len(pw) == 0 {
		return errors.New("password must not be empty")
	}
	cfg.ReaperPassword = string(pw)
	return nil
}
