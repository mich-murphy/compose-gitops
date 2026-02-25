package compose

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDiscoverProjects_Empty(t *testing.T) {
	dir := t.TempDir()
	got, err := DiscoverProjects(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

func TestDiscoverProjects_FindsComposeFiles(t *testing.T) {
	dir := t.TempDir()
	// project a with compose.yml
	projA := filepath.Join(dir, "projA")
	if err := os.MkdirAll(projA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projA, "compose.yml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	// project b with docker-compose.yaml
	projB := filepath.Join(dir, "projB")
	if err := os.MkdirAll(projB, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projB, "docker-compose.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	// nested
	nested := filepath.Join(dir, "nested", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "compose.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	// dir without compose - skip
	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := DiscoverProjects(dir)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	want := []string{"nested/deep", "projA", "projB"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestResolveComposeFile(t *testing.T) {
	dir := t.TempDir()
	if ResolveComposeFile(dir) != "" {
		t.Error("empty dir should return empty")
	}
	if err := os.WriteFile(filepath.Join(dir, "compose.yml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ResolveComposeFile(dir); got == "" {
		t.Error("expected path to compose.yml")
	}
}
