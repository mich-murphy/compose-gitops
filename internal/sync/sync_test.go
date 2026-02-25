package sync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/michael/dev/compose-gitops/internal/compose"
	"github.com/michael/dev/compose-gitops/internal/config"
)

// MockRunner records calls for testing.
type MockRunner struct {
	PullCalls  []string
	UpCalls   []string
	DownCalls []string
}

func (m *MockRunner) Pull(ctx context.Context, projectDir string) error {
	m.PullCalls = append(m.PullCalls, projectDir)
	return nil
}

func (m *MockRunner) Up(ctx context.Context, projectDir string) error {
	m.UpCalls = append(m.UpCalls, projectDir)
	return nil
}

func (m *MockRunner) Down(ctx context.Context, projectDir string) error {
	m.DownCalls = append(m.DownCalls, projectDir)
	return nil
}

func TestResolveProjectList_NoOverride(t *testing.T) {
	cfg := &config.Config{}
	discovered := []string{"a", "b"}
	got := resolveProjectList(cfg, discovered)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v", got)
	}
}

func TestResolveProjectList_OverrideReplaces(t *testing.T) {
	cfg := &config.Config{ProjectsOverride: []string{"x", "y"}, AppendProjects: false}
	discovered := []string{"a", "b"}
	got := resolveProjectList(cfg, discovered)
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Errorf("got %v", got)
	}
}

func TestResolveProjectList_Append(t *testing.T) {
	cfg := &config.Config{ProjectsOverride: []string{"x"}, AppendProjects: true}
	discovered := []string{"a", "b"}
	got := resolveProjectList(cfg, discovered)
	if len(got) != 3 {
		t.Errorf("got %v", got)
	}
}

func TestSync_IntegrationSkipIgnored(t *testing.T) {
	// This test requires git and a real repo; skip in short mode and just test resolveProjectList above.
	// For a full integration test we could use a local file:// repo.
	t.Skip("integration test requires git; unit tests cover resolveProjectList and state")
}

func TestStateRoundtrip(t *testing.T) {
	dir := t.TempDir()
	list := []string{"proj1", "proj2"}
	if err := WriteLastDeployed(dir, list); err != nil {
		t.Fatal(err)
	}
	got, err := ReadLastDeployed(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "proj1" || got[1] != "proj2" {
		t.Errorf("got %v", got)
	}
}

func TestComposeProjectDirs(t *testing.T) {
	got := compose.ComposeProjectDirs("/repo", []string{"a", "b"})
	want := []string{"/repo/a", "/repo/b"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v", got)
	}
}

func TestEnsureWorkDirExists(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	if err := EnsureWorkDirExists(dir); err != nil {
		t.Fatal(err)
	}
	// already exists
	if err := EnsureWorkDirExists(dir); err != nil {
		t.Fatal(err)
	}
}
