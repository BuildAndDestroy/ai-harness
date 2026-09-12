package harness

import (
	"strings"
	"testing"
)

func TestBuildPromptOmitsPasswordAndListsSkills(t *testing.T) {
	cfg := Config{
		ReaperURL:      "https://c2.example.com:8443",
		ReaperUsername: "op1",
		ReaperPassword: "super-secret-value",
		Client:         "Acme Corp",
		Engagement:     "acme-2026-q3",
		Objectives:     []string{"Get domain admin", "Reach the finance share"},
	}

	prompt, err := BuildPrompt(cfg)
	if err != nil {
		t.Fatalf("BuildPrompt() error: %v", err)
	}

	if strings.Contains(prompt, cfg.ReaperPassword) {
		t.Fatal("BuildPrompt() must never embed the password in the prompt text")
	}
	if !strings.Contains(prompt, "$REAPER_PASSWORD") {
		t.Error("prompt should tell the agent to read the password from the environment")
	}

	for _, skill := range []string{
		"red-team-operator",
		"reaperc2-operator",
		"exploit-development",
		"purple-team-atomic-tests",
		"harness-report",
	} {
		if !strings.Contains(prompt, skill) {
			t.Errorf("prompt missing reference to skill %q", skill)
		}
	}

	for _, want := range []string{
		cfg.Client, cfg.Engagement, cfg.ReaperURL, cfg.ReaperUsername,
		"Get domain admin", "Reach the finance share",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing expected content %q", want)
		}
	}
}
