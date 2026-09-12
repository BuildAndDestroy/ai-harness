package harness

import (
	"errors"
	"testing"
)

func TestLaunchClaudeNotFound(t *testing.T) {
	// An empty temp dir on PATH guarantees `claude` can't be resolved.
	t.Setenv("PATH", t.TempDir())

	err := Launch(Config{}, "prompt")
	if !errors.Is(err, ErrClaudeNotFound) {
		t.Fatalf("Launch() error = %v, want ErrClaudeNotFound", err)
	}
}
