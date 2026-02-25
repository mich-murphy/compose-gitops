package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds application configuration from environment variables.
type Config struct {
	// Repository URL to clone/fetch (required).
	Repository string
	// Branch to track (default: main).
	Branch string
	// Working directory for git checkout and compose projects (required).
	WorkDir string
	// PollInterval is the interval for polling when enabled (e.g. "5m"). Zero disables polling.
	PollInterval time.Duration
	// WebhookSecret is optional; if set, webhook requests must include this secret (e.g. in header or query).
	WebhookSecret string
	// IgnoreProjects is a list of project names to never deploy.
	IgnoreProjects []string
	// ProjectsOverride if non-empty, replaces auto-discovered projects. Comma-separated list.
	ProjectsOverride []string
	// AppendProjects if true, ProjectsOverride is appended to auto-discovered list; otherwise replaces.
	AppendProjects bool
	// PullProjectImages when true, run docker compose pull before deploy for each project.
	PullProjectImages bool
	// PruneProjects when true, remove projects that no longer exist in the repository.
	PruneProjects bool
	// HTTPListenAddr for the webhook server (e.g. ":8080").
	HTTPListenAddr string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	repo := os.Getenv("COMPOSE_GITOPS_REPOSITORY")
	if repo == "" {
		return nil, fmt.Errorf("COMPOSE_GITOPS_REPOSITORY is required")
	}
	workDir := os.Getenv("COMPOSE_GITOPS_WORK_DIR")
	if workDir == "" {
		return nil, fmt.Errorf("COMPOSE_GITOPS_WORK_DIR is required")
	}
	branch := getEnv("COMPOSE_GITOPS_BRANCH", "main")
	pollStr := os.Getenv("COMPOSE_GITOPS_POLL_INTERVAL")
	var pollInterval time.Duration
	if pollStr != "" {
		d, err := time.ParseDuration(pollStr)
		if err != nil {
			return nil, fmt.Errorf("COMPOSE_GITOPS_POLL_INTERVAL: %w", err)
		}
		pollInterval = d
	}
	ignore := sliceFromEnv("COMPOSE_GITOPS_IGNORE_PROJECTS")
	override := sliceFromEnv("COMPOSE_GITOPS_PROJECTS")
	appendProjects := getEnv("COMPOSE_GITOPS_APPEND_PROJECTS", "false") == "true" || getEnv("COMPOSE_GITOPS_APPEND_PROJECTS", "false") == "1"
	pullImages := getEnv("COMPOSE_GITOPS_PULL_IMAGES", "true") == "true" || getEnv("COMPOSE_GITOPS_PULL_IMAGES", "true") == "1"
	prune := getEnv("COMPOSE_GITOPS_PRUNE_PROJECTS", "false") == "true" || getEnv("COMPOSE_GITOPS_PRUNE_PROJECTS", "false") == "1"
	listenAddr := getEnv("COMPOSE_GITOPS_HTTP_ADDR", ":8080")
	webhookSecret := os.Getenv("COMPOSE_GITOPS_WEBHOOK_SECRET")

	return &Config{
		Repository:        repo,
		Branch:            branch,
		WorkDir:           workDir,
		PollInterval:      pollInterval,
		WebhookSecret:     webhookSecret,
		IgnoreProjects:    ignore,
		ProjectsOverride:  override,
		AppendProjects:     appendProjects,
		PullProjectImages: pullImages,
		PruneProjects:     prune,
		HTTPListenAddr:    listenAddr,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func sliceFromEnv(key string) []string {
	v := os.Getenv(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// IsIgnored returns true if the project name is in the ignore list.
func (c *Config) IsIgnored(project string) bool {
	for _, ig := range c.IgnoreProjects {
		if ig == project {
			return true
		}
	}
	return false
}

// PollEnabled returns true if polling is configured.
func (c *Config) PollEnabled() bool {
	return c.PollInterval > 0
}

// Validate checks configuration consistency.
func (c *Config) Validate() error {
	if c.Repository == "" {
		return fmt.Errorf("repository is required")
	}
	if c.WorkDir == "" {
		return fmt.Errorf("work_dir is required")
	}
	if c.PollInterval < 0 {
		return fmt.Errorf("poll_interval must be non-negative")
	}
	return nil
}

// IntEnv reads an optional integer from the environment (for tests or future use).
func IntEnv(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
