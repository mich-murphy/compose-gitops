package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/michael/dev/compose-gitops/internal/config"
	syncpkg "github.com/michael/dev/compose-gitops/internal/sync"
)

// Server runs the HTTP webhook server and optional polling loop.
type Server struct {
	cfg    *config.Config
	runner syncpkg.Runner
}

// New creates a new Server.
func New(cfg *config.Config, runner syncpkg.Runner) *Server {
	return &Server{cfg: cfg, runner: runner}
}

// Run starts the HTTP server and optional poller. It blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if err := syncpkg.EnsureWorkDirExists(s.cfg.WorkDir); err != nil {
		return err
	}
	if s.cfg.PollEnabled() {
		go s.runPoller(ctx)
	}
	return s.listenAndServe(ctx)
}

func (s *Server) listenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/sync", s.handleSync)
	mux.HandleFunc("/webhook", s.handleWebhook)
	srv := &http.Server{Addr: s.cfg.HTTPListenAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	slog.Info("http server listening", "addr", s.cfg.HTTPListenAddr)
	return srv.ListenAndServe()
}

func (s *Server) runPoller(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.triggerSync(context.Background())
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.verifySecret(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	result, err := s.triggerSync(r.Context())
	if err != nil {
		slog.Error("sync failed", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.verifySecret(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Run sync in background so we can respond quickly to GitHub (avoid timeout).
	go s.triggerSync(context.Background())
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("accepted"))
}

func (s *Server) verifySecret(r *http.Request) bool {
	got := r.Header.Get("X-Webhook-Secret")
	return s.cfg.WebhookSecret != "" && subtle.ConstantTimeCompare([]byte(got), []byte(s.cfg.WebhookSecret)) == 1
}

var syncMu sync.Mutex

func (s *Server) triggerSync(ctx context.Context) (*syncpkg.SyncResult, error) {
	syncMu.Lock()
	defer syncMu.Unlock()
	return syncpkg.Sync(ctx, s.cfg, s.runner)
}
