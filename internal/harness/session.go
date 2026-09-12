package harness

import (
	"os"
	"path/filepath"
)

// WriteSession writes the rendered prompt to <dir>/<engagement>.md with
// restrictive permissions (dir 0700, file 0600), creating dir if needed, and
// returns the file's path. Session files hold client names and objectives but
// never credentials.
func WriteSession(dir, engagement, content string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return "", err
	}

	path := filepath.Join(dir, engagement+".md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", err
	}
	return path, nil
}
