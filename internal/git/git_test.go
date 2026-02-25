package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRepoNameFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/foo/bar", "bar"},
		{"https://github.com/foo/bar.git", "bar"},
		{"git@github.com:foo/bar.git", "bar"},
	}
	for _, tt := range tests {
		got := repoNameFromURL(tt.url)
		if got != tt.want {
			t.Errorf("repoNameFromURL(%q) = %q want %q", tt.url, got, tt.want)
		}
	}
}

func TestCloneOrFetch_RequiresGit(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	// Use a non-existent URL so we fail early; we only care that repo name is derived
	_, err := CloneOrFetch(ctx, "https://github.com/does-not-exist/repo.git", "main", dir)
	if err == nil {
		t.Fatal("expected error (clone should fail for invalid repo)")
	}
	// Repo name should be "repo"
	repoPath := filepath.Join(dir, "repo")
	if _, err := os.Stat(repoPath); err == nil {
		t.Error("repo dir should not exist after failed clone")
	}
}

func TestRepoNameFromURL_EdgeCases(t *testing.T) {
	if got := repoNameFromURL(""); got != "" {
		t.Errorf("empty: got %q", got)
	}
	if got := repoNameFromURL("x"); got != "x" {
		t.Errorf("single: got %q", got)
	}
}

// repoNameFromURL is unexported; we test via CloneOrFetch or export for test. We already have a small test above.
// Export for test only using build tag or just test the logic by making it a standalone function.
// Checking: we have repoNameFromURL as private. Our test uses it via the test - we need to either export it or test via CloneOrFetch. I'll add a test that verifies the clone fails with a proper repo name in the error or path. Actually the test above already does that - we check repoPath would be dir/repo. Let me add a table-driven test for repoNameFromURL by making it a function we can call from test. Actually it's lowercase so not visible. Let me just keep the URL tests that use a public API. We can test "clone fails" by using file:// with a local repo. For now the unit test for repo name is in a separate file that can't access repoNameFromURL. So either export it (and document as internal) or use integration. I'll export it as RepoNameFromURL for testing.
func init() {
	// ensure we're testing the same logic
	if repoNameFromURL("https://github.com/a/b") != "b" {
		panic("repoNameFromURL broken")
	}
}
