package web

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

//go:embed ui.html
var uiHTML []byte

// JobConfig holds the registration parameters submitted from the web UI.
type JobConfig struct {
	Total               int    `json:"total"`
	Workers             int    `json:"workers"`
	Output              string `json:"output"`
	Domain              string `json:"domain"`
	Password            string `json:"password"`
	Proxy               string `json:"proxy"`
	Pacing              string `json:"pacing"`
	CloudflareMailURL   string `json:"cloudflareMailUrl"`
	CloudflareMailToken string `json:"cloudflareMailToken"`
	ViOTPToken          string `json:"viOTPToken"`
	ViOTPServiceID      int    `json:"viOTPServiceID"`
	CodexEnabled        bool   `json:"codexEnabled"`
	CodexOutput         string `json:"codexOutput"`
	PanelOutputDir      string `json:"panelOutputDir"`
}

// JobResult holds the outcome broadcast when a job finishes.
type JobResult struct {
	Success    int64  `json:"success"`
	Failed     int64  `json:"failed"`
	Target     int    `json:"target"`
	StopReason string `json:"stopReason"`
}

// RunFunc executes a batch registration job, writing progress lines to w.
// It must respect ctx cancellation.
type RunFunc func(ctx context.Context, cfg JobConfig, w io.Writer) (JobResult, error)

// Server hosts the web UI and manages the lifecycle of a single registration job.
type Server struct {
	port         int
	run          RunFunc
	broker       *SSEBroker
	sessionToken string
	mu           sync.Mutex
	cancel       context.CancelFunc // non-nil when a job is running
}

// NewServer creates a Server on the given port using run for job execution.
func NewServer(port int, run RunFunc) *Server {
	return &Server{
		port:         port,
		run:          run,
		broker:       NewSSEBroker(),
		sessionToken: newSessionToken(),
	}
}

// Start starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveUI)
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/stop", s.handleStop)
	mux.HandleFunc("/api/events", s.handleEvents)

	srv := &http.Server{Addr: s.listenAddress(), Handler: mux}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(shutCtx) //nolint:errcheck
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) serveUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "web_session_token",
		Value:    s.sessionToken,
		Path:     "/api/events",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := bytes.ReplaceAll(uiHTML, []byte("__SESSION_TOKEN__"), []byte(s.sessionToken))
	w.Write(page) //nolint:errcheck
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.validSessionToken(r) {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var cfg JobConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		jsonErr(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		jsonErr(w, "a job is already running", http.StatusConflict)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	go s.execJob(ctx, cfg)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"ok":true}`)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.validSessionToken(r) {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"ok":true}`)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if !s.validSessionToken(r) {
		jsonErr(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsub := s.broker.Subscribe()
	defer unsub()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) execJob(ctx context.Context, cfg JobConfig) {
	defer func() {
		s.mu.Lock()
		if s.cancel != nil {
			s.cancel()
			s.cancel = nil
		}
		s.mu.Unlock()
	}()

	result, err := s.run(ctx, cfg, s.broker.LineWriter())
	if err != nil {
		data, _ := json.Marshal(map[string]string{"type": "error", "message": "job failed"})
		s.broker.Send(string(data))
		return
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type":       "done",
		"success":    result.Success,
		"failed":     result.Failed,
		"target":     result.Target,
		"stopReason": result.StopReason,
	})
	s.broker.Send(string(data))
}

func (s *Server) listenAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

func (s *Server) validSessionToken(r *http.Request) bool {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		cookie, err := r.Cookie("web_session_token")
		if err == nil {
			token = cookie.Value
		}
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.sessionToken)) == 1
}

func newSessionToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("generate web session token: %w", err))
	}
	return hex.EncodeToString(buf)
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(data) //nolint:errcheck
}
