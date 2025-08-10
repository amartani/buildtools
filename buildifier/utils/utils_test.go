package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestExpandDirectoriesRespectBazelignore(t *testing.T) {
	tmp := t.TempDir()
	// workspace files
	if err := os.WriteFile(filepath.Join(tmp, "WORKSPACE"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// create .bazelignore
	ignore := []byte("ignored\na/ignored\n")
	if err := os.WriteFile(filepath.Join(tmp, ".bazelignore"), ignore, 0644); err != nil {
		t.Fatal(err)
	}
	// create directories and BUILD files
	paths := []string{
		filepath.Join(tmp, "BUILD"),
		filepath.Join(tmp, "a", "BUILD"),
		filepath.Join(tmp, "a", "ignored", "BUILD"),
		filepath.Join(tmp, "ignored", "BUILD"),
		filepath.Join(tmp, "b", "BUILD"),
	}
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{tmp}
	files, err := ExpandDirectories(&args, true)
	if err != nil {
		t.Fatalf("ExpandDirectories returned error: %v", err)
	}
	sort.Strings(files)
	want := []string{
		filepath.Join(tmp, "BUILD"),
		filepath.Join(tmp, "WORKSPACE"),
		filepath.Join(tmp, "a", "BUILD"),
		filepath.Join(tmp, "b", "BUILD"),
	}
	sort.Strings(want)
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected files: got %v want %v", files, want)
	}
	// When respect_bazelignore is false, ignored files should appear.
	files, err = ExpandDirectories(&args, false)
	if err != nil {
		t.Fatalf("ExpandDirectories returned error: %v", err)
	}
	sort.Strings(files)
	want = append(want, filepath.Join(tmp, "a", "ignored", "BUILD"), filepath.Join(tmp, "ignored", "BUILD"))
	sort.Strings(want)
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected files without respect: got %v want %v", files, want)
	}
}
