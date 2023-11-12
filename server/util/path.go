package util

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Abs(filepath.Join(home, path[1:]))
	}
	return path, nil
}
