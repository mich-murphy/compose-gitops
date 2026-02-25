package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/michael/dev/compose-gitops/internal/config"
)

// mockRunner is a no-op Runner for tests that do not run real compose operations.
type mockRunner struct{}

func (mockRunner) Pull(context.Context, string) error  { return nil }
func (mockRunner) Up(context.Context, string) error     { return nil }
func (mockRunner) Down(context.Context, string) error   { return nil }

func TestHandleHealth(t *testing.T) {
	cfg := &config.Config{Repository: "x", WorkDir: t.TempDir()}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("code: got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("body: got %q", w.Body.String())
	}
}

func TestHandleSync_MethodNotAllowed(t *testing.T) {
	cfg := &config.Config{Repository: "x", WorkDir: t.TempDir()}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodGet, "/sync", nil)
	w := httptest.NewRecorder()
	srv.handleSync(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("code: got %d", w.Code)
	}
}

func TestHandleSync_UnauthorizedWhenSecretSet(t *testing.T) {
	cfg := &config.Config{Repository: "x", WorkDir: t.TempDir(), WebhookSecret: "secret123"}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	w := httptest.NewRecorder()
	srv.handleSync(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("code: got %d", w.Code)
	}
}

func TestVerifySecret_EmptyAllowed(t *testing.T) {
	cfg := &config.Config{}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	if !srv.verifySecret(req) {
		t.Error("empty secret should allow")
	}
}

func TestVerifySecret_Header(t *testing.T) {
	cfg := &config.Config{WebhookSecret: "abc"}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set("X-Webhook-Secret", "abc")
	if !srv.verifySecret(req) {
		t.Error("header secret should match")
	}
	req.Header.Set("X-Webhook-Secret", "wrong")
	if srv.verifySecret(req) {
		t.Error("wrong secret should not match")
	}
}
