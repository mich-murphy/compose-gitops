package config

import (
	"os"
	"testing"
)

func TestLoad_RequiredMissing(t *testing.T) {
	os.Clearenv()
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when COMPOSE_GITOPS_REPOSITORY and WORK_DIR missing")
	}
}

func TestLoad_RepositoryRequired(t *testing.T) {
	os.Clearenv()
	os.Setenv("COMPOSE_GITOPS_WORK_DIR", "/tmp/work")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when repository missing")
	}
}

func TestLoad_WorkDirRequired(t *testing.T) {
	os.Clearenv()
	os.Setenv("COMPOSE_GITOPS_REPOSITORY", "https://github.com/foo/bar")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when work dir missing")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	os.Setenv("COMPOSE_GITOPS_REPOSITORY", "https://github.com/foo/bar")
	os.Setenv("COMPOSE_GITOPS_WORK_DIR", "/tmp/work")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Branch != "main" {
		t.Errorf("branch: got %q", cfg.Branch)
	}
	if cfg.HTTPListenAddr != ":8080" {
		t.Errorf("listen addr: got %q", cfg.HTTPListenAddr)
	}
	if cfg.PollInterval != 0 {
		t.Errorf("poll interval: got %v", cfg.PollInterval)
	}
}

func TestLoad_IgnoreProjects(t *testing.T) {
	os.Clearenv()
	os.Setenv("COMPOSE_GITOPS_REPOSITORY", "https://github.com/foo/bar")
	os.Setenv("COMPOSE_GITOPS_WORK_DIR", "/tmp/work")
	os.Setenv("COMPOSE_GITOPS_IGNORE_PROJECTS", "a, b , c")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.IgnoreProjects) != 3 {
		t.Errorf("ignore projects: got %v", cfg.IgnoreProjects)
	}
}

func TestIsIgnored(t *testing.T) {
	cfg := &Config{IgnoreProjects: []string{"skip1", "skip2"}}
	if !cfg.IsIgnored("skip1") {
		t.Error("skip1 should be ignored")
	}
	if cfg.IsIgnored("other") {
		t.Error("other should not be ignored")
	}
}

func TestValidate(t *testing.T) {
	if err := (&Config{}).Validate(); err == nil {
		t.Error("empty config should be invalid")
	}
	if err := (&Config{Repository: "x", WorkDir: "y"}).Validate(); err != nil {
		t.Errorf("valid config: %v", err)
	}
}
