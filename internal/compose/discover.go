package compose

import (
	"os"
	"path/filepath"
	"sort"
)

// ComposeFilenames are the file names we treat as a compose project file.
var ComposeFilenames = []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}

// DiscoverProjects finds all directories under repoRoot that contain a compose file.
// Returns project names (directory names) sorted for deterministic behaviour.
func DiscoverProjects(repoRoot string) ([]string, error) {
	var projects []string
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			return nil
		}
		dirName := filepath.Base(path)
		if dirName == ".git" {
			return filepath.SkipDir
		}
		for _, name := range ComposeFilenames {
			f := filepath.Join(path, name)
			if _, err := os.Stat(f); err == nil {
				rel, _ := filepath.Rel(repoRoot, path)
				if rel == "." {
					rel = filepath.Base(repoRoot)
				}
				projects = append(projects, rel)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(projects)
	return projects, nil
}

// ProjectPath returns the absolute path to a project directory inside repoRoot.
func ProjectPath(repoRoot, project string) string {
	return filepath.Join(repoRoot, project)
}

// ResolveComposeFile returns the path to the compose file in the project dir, or empty if not found.
func ResolveComposeFile(projectDir string) string {
	for _, name := range ComposeFilenames {
		p := filepath.Join(projectDir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
