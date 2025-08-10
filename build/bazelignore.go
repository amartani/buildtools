/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// GetIgnoredPrefixes returns a list of ignored prefixes from the .bazelignore file in the root directory.
// It returns an empty list if the file does not exist.
func GetIgnoredPrefixes(rootDir string) []string {
	bazelignorePath := filepath.Join(rootDir, ".bazelignore")
	ignoredPaths := []string{}

	data, err := os.ReadFile(bazelignorePath)
	if err != nil {
		return ignoredPaths
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines, comments, and absolute paths
		// Bazel will error out if there are any absolute paths in the .bazelignore file.
		if line == "" || strings.HasPrefix(line, "#") || path.IsAbs(line) {
			continue
		}

		ignoredPaths = append(ignoredPaths, path.Clean(line))
	}

	return ignoredPaths
}

// ShouldIgnorePath returns true if the path should be ignored based on the list of ignored prefixes.
func ShouldIgnorePath(path string, rootDir string, ignoredPrefixes []string) bool {
	if len(ignoredPrefixes) == 0 {
		return false
	}

	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false
	}
	// Normalize path separators to forward slashes
	rel = filepath.ToSlash(rel)

	for _, prefix := range ignoredPrefixes {
		// Check if the path exactly matches the prefix or if it's a subdirectory of the prefix.
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}

// BuildFileNames is a list of possible names for BUILD files.
var BuildFileNames = [...]string{"BUILD.bazel", "BUILD", "BUCK"}

// FindBuildFiles returns all "BUILD" files in that subtree recursively.
// ignoredPrefixes are path prefixes to ignore (if a path matches any of these prefixes,
// it will be skipped along with its subdirectories).
func FindBuildFiles(rootDir string, ignoredPrefixes []string) []string {
	var buildFiles []string
	searchDirs := []string{rootDir}

	for len(searchDirs) != 0 {
		lastIndex := len(searchDirs) - 1
		dir := searchDirs[lastIndex]
		searchDirs = searchDirs[:lastIndex]

		dirFiles, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, dirFile := range dirFiles {
			fullPath := filepath.Join(dir, dirFile.Name())

			if ShouldIgnorePath(fullPath, rootDir, ignoredPrefixes) {
				continue
			}

			if dirFile.IsDir() {
				searchDirs = append(searchDirs, fullPath)
			} else {
				for _, buildFileName := range BuildFileNames {
					if dirFile.Name() == buildFileName {
						buildFiles = append(buildFiles, fullPath)
					}
				}
			}
		}
	}

	return buildFiles
}
