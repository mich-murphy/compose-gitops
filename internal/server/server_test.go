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

func (mockRunner) Pull(context.Context, string) error { return nil }
func (mockRunner) Up(context.Context, string) error   { return nil }
func (mockRunner) Down(context.Context, string) error { return nil }

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

func TestHandleSync_RejectsMissingAndWrongSecret(t *testing.T) {
	cfg := &config.Config{Repository: "x", WorkDir: t.TempDir(), WebhookSecret: "secret123"}
	srv := New(cfg, mockRunner{})
	for name, header := range map[string]string{"missing": "", "wrong": "nope"} {
		req := httptest.NewRequest(http.MethodPost, "/sync", nil)
		if header != "" {
			req.Header.Set("X-Webhook-Secret", header)
		}
		w := httptest.NewRecorder()
		srv.handleSync(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: code: got %d", name, w.Code)
		}
	}
}

func TestHandleWebhook_AcceptsCorrectHeader(t *testing.T) {
	cfg := &config.Config{Repository: "x", WorkDir: t.TempDir(), WebhookSecret: "secret123"}
	srv := New(cfg, mockRunner{})
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	req.Header.Set("X-Webhook-Secret", "secret123")
	w := httptest.NewRecorder()
	srv.handleWebhook(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("code: got %d", w.Code)
	}
}

func TestVerifySecret(t *testing.T) {
	cases := []struct {
		name   string
		secret string
		target string
		header string
		want   bool
	}{
		{"correct header", "abc", "/sync", "abc", true},
		{"wrong header", "abc", "/sync", "wrong", false},
		{"no header", "abc", "/sync", "", false},
		{"query param only", "abc", "/sync?secret=abc", "", false},
		{"empty configured secret, no header", "", "/sync", "", false},
	}
	for _, c := range cases {
		srv := New(&config.Config{WebhookSecret: c.secret}, mockRunner{})
		req := httptest.NewRequest(http.MethodPost, c.target, nil)
		if c.header != "" {
			req.Header.Set("X-Webhook-Secret", c.header)
		}
		if got := srv.verifySecret(req); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
