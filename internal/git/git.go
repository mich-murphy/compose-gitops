package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOrFetch performs git clone (if dir is empty) or fetch + checkout/merge for the given repo and branch.
// workDir is the parent directory; the repo is cloned into workDir/repoName where repoName is derived from the repo URL.
func CloneOrFetch(ctx context.Context, repoURL, branch, workDir string) (repoPath string, err error) {
	repoName := repoNameFromURL(repoURL)
	if repoName == "" {
		return "", fmt.Errorf("could not derive repo name from URL: %s", repoURL)
	}
	repoPath = filepath.Join(workDir, repoName)
	gitDir := filepath.Join(repoPath, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone
		if err := os.MkdirAll(workDir, 0o755); err != nil {
			return "", fmt.Errorf("mkdir work dir: %w", err)
		}
		cmd := exec.CommandContext(ctx, "git", "clone", "--branch", branch, "--single-branch", "--depth", "1", repoURL, repoPath)
		cmd.Dir = workDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git clone: %w: %s", err, string(out))
		}
		return repoPath, nil
	}

	// Fetch and reset to origin/branch (merge semantics: update to latest remote)
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin", branch)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git fetch: %w: %s", err, string(out))
	}
	cmd = exec.CommandContext(ctx, "git", "reset", "--hard", "origin/"+branch)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git reset: %w: %s", err, string(out))
	}
	cmd = exec.CommandContext(ctx, "git", "clean", "-fd")
	cmd.Dir = repoPath
	_ = cmd.Run() // ignore clean errors
	return repoPath, nil
}

func repoNameFromURL(url string) string {
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
