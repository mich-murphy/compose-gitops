package sync

import (
	"os"
	"path/filepath"
	"strings"
)

const stateFileName = ".compose-gitops-deployed"

// ReadLastDeployed reads the list of project names we last deployed (for prune).
// Returns relative project names (e.g. "plex", "docker/arrs").
func ReadLastDeployed(workDir string) ([]string, error) {
	p := filepath.Join(workDir, stateFileName)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	line := strings.TrimSpace(string(b))
	if line == "" {
		return nil, nil
	}
	return strings.Split(line, "\n"), nil
}

// WriteLastDeployed records the list of project names we deployed this run.
func WriteLastDeployed(workDir string, projects []string) error {
	p := filepath.Join(workDir, stateFileName)
	return os.WriteFile(p, []byte(strings.Join(projects, "\n")), 0o600)
}
