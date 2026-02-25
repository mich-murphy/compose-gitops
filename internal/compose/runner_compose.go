package compose

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
)

// ComposeRunner runs docker compose operations via the official Docker Compose Go API.
// It implements Runner by loading the project from disk and calling the compose service.
type ComposeRunner struct {
	api api.Compose
}

// NewComposeRunner creates a Runner that uses the Docker Compose library.
// It initializes the Docker CLI (using default socket and config) and the Compose service.
func NewComposeRunner() (Runner, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("docker cli: %w", err)
	}
	if err := dockerCli.Initialize(flags.NewClientOptions()); err != nil {
		return nil, fmt.Errorf("docker cli init: %w", err)
	}
	svc, err := compose.NewComposeService(dockerCli, compose.WithPrompt(compose.AlwaysOkPrompt()))
	if err != nil {
		return nil, fmt.Errorf("compose service: %w", err)
	}
	return &ComposeRunner{api: svc}, nil
}

// loadProject loads the compose project from projectDir using the first found compose file.
func (r *ComposeRunner) loadProject(ctx context.Context, projectDir string) (*types.Project, error) {
	composeFile := ResolveComposeFile(projectDir)
	if composeFile == "" {
		return nil, fmt.Errorf("no compose file found in %s", projectDir)
	}
	opts := api.ProjectLoadOptions{
		WorkingDir:  projectDir,
		ConfigPaths: []string{filepath.Base(composeFile)},
	}
	project, err := r.api.LoadProject(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("load project: %w", err)
	}
	return project, nil
}

// Pull runs compose pull for the project in projectDir.
func (r *ComposeRunner) Pull(ctx context.Context, projectDir string) error {
	project, err := r.loadProject(ctx, projectDir)
	if err != nil {
		return err
	}
	return r.api.Pull(ctx, project, api.PullOptions{
		Quiet:          true,
		IgnoreFailures: false,
	})
}

// Up runs compose up -d for the project in projectDir.
func (r *ComposeRunner) Up(ctx context.Context, projectDir string) error {
	project, err := r.loadProject(ctx, projectDir)
	if err != nil {
		return err
	}
	return r.api.Up(ctx, project, api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			Recreate:      api.RecreateDiverged,
		},
		Start: api.StartOptions{},
	})
}

// Down runs compose down for the project in projectDir.
func (r *ComposeRunner) Down(ctx context.Context, projectDir string) error {
	project, err := r.loadProject(ctx, projectDir)
	if err != nil {
		return err
	}
	return r.api.Down(ctx, project.Name, api.DownOptions{
		RemoveOrphans: true,
	})
}
