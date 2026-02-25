package compose

import (
	"context"
	"path/filepath"
)

// Runner runs docker compose operations (pull, up, down).
// Production code uses NewComposeRunner() which uses the Docker Compose Go API.
// Tests can use a mock implementation.
type Runner interface {
	Pull(ctx context.Context, projectDir string) error
	Up(ctx context.Context, projectDir string) error
	Down(ctx context.Context, projectDir string) error
}

// DeployOptions control deploy behaviour for a single project.
type DeployOptions struct {
	PullImages bool
}

// Deploy pulls (if requested) and brings up the project.
func Deploy(ctx context.Context, r Runner, projectDir string, opts DeployOptions) error {
	if opts.PullImages {
		if err := r.Pull(ctx, projectDir); err != nil {
			return err
		}
	}
	return r.Up(ctx, projectDir)
}

// Remove brings down and removes the project (for prune).
func Remove(ctx context.Context, r Runner, projectDir string) error {
	return r.Down(ctx, projectDir)
}

// ProjectDirExists returns true if the path exists and contains a compose file.
func ProjectDirExists(projectDir string) bool {
	return ResolveComposeFile(projectDir) != ""
}

// ComposeProjectDirs returns the list of project directories under repoRoot (absolute paths).
func ComposeProjectDirs(repoRoot string, projects []string) []string {
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, filepath.Join(repoRoot, p))
	}
	return out
}
