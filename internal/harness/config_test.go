package harness

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clearReaperEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"REAPER_URL", "REAPER_C2_URL", "REAPER_USERNAME", "REAPER_PASSWORD", "REAPER_ENGAGEMENT"} {
		t.Setenv(k, "")
	}
}

func TestParseArgsAndValidate(t *testing.T) {
	clearReaperEnv(t)

	cases := []struct {
		name    string
		args    []string
		wantErr string // substring expected in Validate() error; "" means no error
	}{
		{
			name:    "missing everything",
			args:    []string{},
			wantErr: "missing required input(s)",
		},
		{
			name:    "missing objective",
			args:    []string{"--reaper-url", "https://c2:8443", "--reaper-c2-url", "https://c2:8080", "--reaper-username", "op1", "--engagement", "acme"},
			wantErr: "--objective",
		},
		{
			name:    "missing engagement",
			args:    []string{"--reaper-url", "https://c2:8443", "--reaper-c2-url", "https://c2:8080", "--reaper-username", "op1", "--objective", "test"},
			wantErr: "--engagement",
		},
		{
			name:    "missing c2 url",
			args:    []string{"--reaper-url", "https://c2:8443", "--reaper-username", "op1", "--engagement", "acme", "--objective", "test"},
			wantErr: "--reaper-c2-url",
		},
		{
			name: "bad scheme",
			args: []string{
				"--reaper-url", "ftp://c2", "--reaper-c2-url", "https://c2:8080",
				"--reaper-username", "op1", "--engagement", "acme",
				"--objective", "test",
			},
			wantErr: "http:// or https://",
		},
		{
			name: "bad c2 scheme",
			args: []string{
				"--reaper-url", "https://c2:8443", "--reaper-c2-url", "ftp://c2:8080",
				"--reaper-username", "op1", "--engagement", "acme",
				"--objective", "test",
			},
			wantErr: "--reaper-c2-url must start with http:// or https://",
		},
		{
			name: "c2 url same as admin url",
			args: []string{
				"--reaper-url", "https://c2.example.com:8443",
				"--reaper-c2-url", "https://c2.example.com:8443/",
				"--reaper-username", "op1",
				"--engagement", "acme",
				"--objective", "test",
			},
			wantErr: "must differ",
		},
		{
			name: "valid",
			args: []string{
				"--reaper-url", "https://c2.example.com:8443",
				"--reaper-c2-url", "https://c2.example.com:8080",
				"--reaper-username", "op1",
				"--engagement", "acme-2026-q3",
				"--objective", "Get domain admin",
				"--objective", "Reach the finance share",
			},
			wantErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer
			cfg, err := ParseArgs(tc.args, &stderr)
			if err != nil {
				t.Fatalf("ParseArgs() unexpected error: %v", err)
			}
			err = cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestParseArgsObjectivesFile(t *testing.T) {
	clearReaperEnv(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "objectives.txt")
	if err := os.WriteFile(path, []byte("First objective\n\nSecond objective\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	cfg, err := ParseArgs([]string{
		"--reaper-url", "https://c2:8443",
		"--reaper-username", "op1",
		"--objective", "Inline objective",
		"--objectives-file", path,
	}, &stderr)
	if err != nil {
		t.Fatalf("ParseArgs() error: %v", err)
	}

	want := []string{"Inline objective", "First objective", "Second objective"}
	if len(cfg.Objectives) != len(want) {
		t.Fatalf("got %d objectives, want %d: %v", len(cfg.Objectives), len(want), cfg.Objectives)
	}
	for i, w := range want {
		if cfg.Objectives[i] != w {
			t.Errorf("objective[%d] = %q, want %q", i, cfg.Objectives[i], w)
		}
	}
}

func TestParseArgsEnvFallback(t *testing.T) {
	clearReaperEnv(t)
	t.Setenv("REAPER_URL", "https://from-env:8443")
	t.Setenv("REAPER_C2_URL", "https://from-env:8080")
	t.Setenv("REAPER_USERNAME", "env-user")
	t.Setenv("REAPER_PASSWORD", "env-pass")
	t.Setenv("REAPER_ENGAGEMENT", "env-eng")

	var stderr bytes.Buffer
	cfg, err := ParseArgs([]string{"--objective", "x"}, &stderr)
	if err != nil {
		t.Fatalf("ParseArgs() error: %v", err)
	}
	if cfg.ReaperURL != "https://from-env:8443" || cfg.ReaperC2URL != "https://from-env:8080" ||
		cfg.ReaperUsername != "env-user" || cfg.ReaperPassword != "env-pass" || cfg.Engagement != "env-eng" {
		t.Fatalf("env fallback not applied: url=%q c2=%q username=%q engagement=%q passwordSet=%v",
			cfg.ReaperURL, cfg.ReaperC2URL, cfg.ReaperUsername, cfg.Engagement, cfg.ReaperPassword != "")
	}

	// An explicit flag still wins over the environment.
	cfg2, err := ParseArgs([]string{"--reaper-username", "flag-user", "--objective", "x"}, &stderr)
	if err != nil {
		t.Fatalf("ParseArgs() error: %v", err)
	}
	if cfg2.ReaperUsername != "flag-user" {
		t.Fatalf("flag should override env, got %q", cfg2.ReaperUsername)
	}
}

func TestFinalizeDefaults(t *testing.T) {
	cfg := &Config{Engagement: "acme-2026-q3"}
	cfg.Finalize()
	if cfg.Engagement != "acme-2026-q3" {
		t.Errorf("Engagement = %q, Finalize must not invent an engagement name", cfg.Engagement)
	}
	if cfg.Client != "<UPDATE ME>" {
		t.Errorf("Client = %q", cfg.Client)
	}
	if cfg.SessionsDir != "sessions" {
		t.Errorf("SessionsDir = %q", cfg.SessionsDir)
	}

	cfg2 := &Config{Engagement: "acme", Client: "Acme", SessionsDir: "custom"}
	cfg2.Finalize()
	if cfg2.Engagement != "acme" || cfg2.Client != "Acme" || cfg2.SessionsDir != "custom" {
		t.Errorf("Finalize overwrote explicit values: engagement=%q client=%q sessionsDir=%q",
			cfg2.Engagement, cfg2.Client, cfg2.SessionsDir)
	}
}

func TestConfigStringRedactsPassword(t *testing.T) {
	cfg := Config{ReaperUsername: "op1", ReaperPassword: "super-secret-value"}

	// Covers both the value and pointer forms, since fmt dispatches to
	// Stringer differently depending on which is passed to %v/%+v.
	for _, formatted := range []string{
		fmt.Sprintf("%v", cfg),
		fmt.Sprintf("%+v", cfg),
		fmt.Sprintf("%v", &cfg),
		fmt.Sprintf("%+v", &cfg),
	} {
		if strings.Contains(formatted, cfg.ReaperPassword) {
			t.Fatalf("Config formatting leaked the password: %s", formatted)
		}
		if !strings.Contains(formatted, "<redacted>") {
			t.Fatalf("Config formatting should note the password is redacted: %s", formatted)
		}
	}
}
