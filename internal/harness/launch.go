package harness

import (
	"errors"
	"os"
	"os/exec"
)

// ErrClaudeNotFound indicates the claude CLI isn't on PATH.
var ErrClaudeNotFound = errors.New("claude CLI not found on PATH")

// Launch runs `claude <prompt>`, injecting the ReaperC2 connection details into
// its process environment only — they are never written to disk. It blocks until
// claude exits; a non-zero exit is returned as *exec.ExitError so callers can
// propagate the exact code.
func Launch(cfg Config, prompt string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return ErrClaudeNotFound
	}

	cmd := exec.Command(bin, prompt)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"REAPER_URL="+cfg.ReaperURL,
		"REAPER_C2_URL="+cfg.ReaperC2URL,
		"REAPER_USERNAME="+cfg.ReaperUsername,
		"REAPER_PASSWORD="+cfg.ReaperPassword,
		"REAPER_ENGAGEMENT="+cfg.Engagement,
	)
	return cmd.Run()
}
