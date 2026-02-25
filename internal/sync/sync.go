package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/michael/dev/compose-gitops/internal/compose"
	"github.com/michael/dev/compose-gitops/internal/config"
	"github.com/michael/dev/compose-gitops/internal/git"
)

// SyncResult holds the outcome of a sync run.
type SyncResult struct {
	RepoPath    string
	Projects    []string
	Deployed    []string
	Skipped     []string
	Pruned      []string
	Errors      []string
}

// Runner is the compose command runner (for tests).
type Runner = compose.Runner

// Sync runs the full flow: git fetch/merge, discover projects, deploy or skip, optionally prune.
func Sync(ctx context.Context, cfg *config.Config, run Runner) (*SyncResult, error) {
	result := &SyncResult{}

	repoPath, err := git.CloneOrFetch(ctx, cfg.Repository, cfg.Branch, cfg.WorkDir)
	if err != nil {
		return result, fmt.Errorf("git: %w", err)
	}
	result.RepoPath = repoPath

	discovered, err := compose.DiscoverProjects(repoPath)
	if err != nil {
		return result, fmt.Errorf("discover projects: %w", err)
	}

	projects := resolveProjectList(cfg, discovered)
	result.Projects = projects

	// Prune: remove projects that were previously deployed but are no longer in the effective list.
	if cfg.PruneProjects {
		lastDeployed, _ := ReadLastDeployed(cfg.WorkDir)
		currentSet := make(map[string]bool)
		for _, p := range projects {
			currentSet[p] = true
		}
		for _, proj := range lastDeployed {
			if currentSet[proj] {
				continue
			}
			projectDir := filepath.Join(repoPath, proj)
			if !compose.ProjectDirExists(projectDir) {
				slog.Info("prune: skip (dir removed from repo)", "project", proj)
				continue
			}
			if err := compose.Remove(ctx, run, projectDir); err != nil {
				slog.Error("prune: remove failed", "project", proj, "error", err)
				result.Errors = append(result.Errors, fmt.Sprintf("prune %s: %v", proj, err))
				continue
			}
			result.Pruned = append(result.Pruned, proj)
			slog.Info("prune: removed", "project", proj)
		}
	}

	for _, proj := range projects {
		if cfg.IsIgnored(proj) {
			result.Skipped = append(result.Skipped, proj)
			slog.Info("sync: skipped (ignored)", "project", proj)
			continue
		}
		projectDir := filepath.Join(repoPath, proj)
		if compose.ResolveComposeFile(projectDir) == "" {
			result.Skipped = append(result.Skipped, proj)
			slog.Info("sync: skipped (no compose file)", "project", proj)
			continue
		}
		opts := compose.DeployOptions{PullImages: cfg.PullProjectImages}
		if err := compose.Deploy(ctx, run, projectDir, opts); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", proj, err))
			slog.Error("sync: deploy failed", "project", proj, "error", err)
			continue
		}
		result.Deployed = append(result.Deployed, proj)
		slog.Info("sync: deployed", "project", proj)
	}

	// Persist deployed list for next run's prune.
	deployedForState := append([]string(nil), result.Deployed...)
	if err := WriteLastDeployed(cfg.WorkDir, deployedForState); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("write state: %v", err))
	}

	return result, nil
}

func resolveProjectList(cfg *config.Config, discovered []string) []string {
	if len(cfg.ProjectsOverride) == 0 {
		return discovered
	}
	if cfg.AppendProjects {
		seen := make(map[string]bool)
		for _, p := range discovered {
			seen[p] = true
		}
		for _, p := range cfg.ProjectsOverride {
			if !seen[p] {
				discovered = append(discovered, p)
				seen[p] = true
			}
		}
		return discovered
	}
	return cfg.ProjectsOverride
}

// EnsureWorkDirExists creates the work dir if it doesn't exist.
func EnsureWorkDirExists(workDir string) error {
	return os.MkdirAll(workDir, 0o755)
}
