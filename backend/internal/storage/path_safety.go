package storage

import (
	"fmt"
	"path"
	"strings"
)

func cleanStoragePath(p string) (string, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	if normalized == "" || path.IsAbs(normalized) {
		return "", fmt.Errorf("invalid path")
	}
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", fmt.Errorf("invalid path")
		}
	}
	clean := path.Clean(normalized)
	if clean == "." || clean == "" {
		return "", fmt.Errorf("invalid path")
	}
	return clean, nil
}

func cleanStoragePrefix(prefix string) (string, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(prefix, "\\", "/"))
	keepTrailingSlash := strings.HasSuffix(normalized, "/")
	clean, err := cleanStoragePath(normalized)
	if err != nil {
		return "", err
	}
	if keepTrailingSlash && !strings.HasSuffix(clean, "/") {
		clean += "/"
	}
	return clean, nil
}
