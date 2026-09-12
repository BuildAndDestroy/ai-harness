// Command harness launches a Claude Code session primed to run one ai-harness
// engagement against ReaperC2, using this project's red-team-operator,
// reaperc2-operator, exploit-development, purple-team-atomic-tests, and
// harness-report skills.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/BuildAndDestroy/ai-harness/internal/harness"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := harness.ParseArgs(os.Args[1:], os.Stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, harness.UsageText)
		return 1
	}

	cfg.Finalize()

	if err := harness.ResolvePassword(cfg, os.Stdin, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	prompt, err := harness.BuildPrompt(*cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error building prompt:", err)
		return 1
	}

	sessionPath, err := harness.WriteSession(cfg.SessionsDir, cfg.Engagement, prompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing session file:", err)
		return 1
	}
	fmt.Fprintln(os.Stderr, "Session prompt written to:", sessionPath)
	fmt.Fprintln(os.Stderr, "(credentials were not written to this file — password lives only in this process's environment)")

	if cfg.DryRun {
		fmt.Fprintln(os.Stderr)
		fmt.Print(prompt)
		return 0
	}

	if err := harness.Launch(*cfg, prompt); err != nil {
		if errors.Is(err, harness.ErrClaudeNotFound) {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, "claude CLI not found on PATH. Prompt is ready at:", sessionPath)
			fmt.Fprintln(os.Stderr, "Export REAPER_URL, REAPER_C2_URL, REAPER_USERNAME, REAPER_PASSWORD, REAPER_ENGAGEMENT yourself, then run:")
			fmt.Fprintf(os.Stderr, "  claude \"$(cat %s)\"\n", sessionPath)
			return 0
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "Error launching claude:", err)
		return 1
	}
	return 0
}
