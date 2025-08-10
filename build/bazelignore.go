package build

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// GetIgnoredPrefixes returns a list of ignored path prefixes from the .bazelignore
// file located in rootDir. If the file doesn't exist or can't be read, it
// returns an empty list.
func GetIgnoredPrefixes(rootDir string) []string {
	bazelignorePath := filepath.Join(rootDir, ".bazelignore")
	data, err := os.ReadFile(bazelignorePath)
	if err != nil {
		return []string{}
	}
	prefixes := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || path.IsAbs(line) {
			continue
		}
		prefixes = append(prefixes, path.Clean(line))
	}
	return prefixes
}

// ShouldIgnorePath reports whether path should be ignored based on the provided
// list of ignored prefixes. The path is checked relative to rootDir.
func ShouldIgnorePath(p, rootDir string, ignoredPrefixes []string) bool {
	if len(ignoredPrefixes) == 0 {
		return false
	}
	rel, err := filepath.Rel(rootDir, p)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	for _, prefix := range ignoredPrefixes {
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}
