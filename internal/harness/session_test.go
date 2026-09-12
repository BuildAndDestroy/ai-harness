package harness

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteSessionPermissionsAndContent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits not meaningful on windows")
	}

	dir := filepath.Join(t.TempDir(), "sessions")
	path, err := WriteSession(dir, "acme-2026-q3", "hello world")
	if err != nil {
		t.Fatalf("WriteSession() error: %v", err)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("session dir perm = %o, want 0700", perm)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fileInfo.Mode().Perm(); perm != 0o600 {
		t.Errorf("session file perm = %o, want 0600", perm)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Errorf("content = %q", data)
	}
}
