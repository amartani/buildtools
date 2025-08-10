package build

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetIgnoredPrefixes(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	tests := []struct {
		name        string
		bazelignore string
		expected    []string
	}{
		{
			name: "valid paths",
			bazelignore: `# Ignore these directories
ignored
a/ignored
b/c/d`,
			expected: []string{"ignored", "a/ignored", "b/c/d"},
		},
		{
			name:        "empty file",
			bazelignore: ` `,
			expected:    []string{},
		},
		{
			name: "only comments",
			bazelignore: `# This is a comment
# Another comment`,
			expected: []string{},
		},
		{
			name: "empty lines",
			bazelignore: `ignored

a/ignored

# comment
b/c/d`,
			expected: []string{"ignored", "a/ignored", "b/c/d"},
		},
		{
			name: "absolute paths",
			bazelignore: `/absolute/path
ignored
/another/absolute/path`,
			expected: []string{"ignored"},
		},
		{
			name:        "no file",
			bazelignore: "",
			expected:    []string{},
		},
		{
			name:        "trailing slash should be normalized",
			bazelignore: `ignored/`,
			expected:    []string{"ignored"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.MkdirTemp(tmp, "")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)

			if tt.bazelignore != "" {
				if err := os.WriteFile(filepath.Join(dir, ".bazelignore"), []byte(tt.bazelignore), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got := GetIgnoredPrefixes(dir)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetIgnoredPrefixes() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldIgnorePath(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	tests := []struct {
		name            string
		path            string
		ignoredPrefixes []string
		want            bool
	}{
		{
			name:            "exact match",
			path:            filepath.Join(tmp, "foo"),
			ignoredPrefixes: []string{"foo"},
			want:            true,
		},
		{
			name:            "subdirectory",
			path:            filepath.Join(tmp, "foo", "bar"),
			ignoredPrefixes: []string{"foo"},
			want:            true,
		},
		{
			name:            "similar prefix but not directory",
			path:            filepath.Join(tmp, "foobar"),
			ignoredPrefixes: []string{"foo"},
			want:            false,
		},
		{
			name:            "matched with multiple prefixes",
			path:            filepath.Join(tmp, "foobar"),
			ignoredPrefixes: []string{"foo2", "foobar", "baz"},
			want:            true,
		},
		{
			name:            "no match",
			path:            filepath.Join(tmp, "bar"),
			ignoredPrefixes: []string{"foo", "baz"},
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldIgnorePath(tt.path, tmp, tt.ignoredPrefixes); got != tt.want {
				t.Errorf("ShouldIgnorePath(%q, %q, %v) = %v, want %v", tt.path, tmp, tt.ignoredPrefixes, got, tt.want)
			}
		})
	}
}
