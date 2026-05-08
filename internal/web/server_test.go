package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerBindsLoopback(t *testing.T) {
	s := NewServer(8899, nil)
	if got := s.listenAddress(); got != "127.0.0.1:8899" {
		t.Fatalf("listenAddress() = %q", got)
	}
}

func TestServeUIInjectsSessionToken(t *testing.T) {
	s := NewServer(8899, nil)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	s.serveUI(w, r)

	body := w.Body.String()
	if strings.Contains(body, "__SESSION_TOKEN__") {
		t.Fatal("session token placeholder was not replaced")
	}
	if !strings.Contains(body, s.sessionToken) {
		t.Fatal("session token missing from UI")
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "web_session_token" || cookies[0].Value != s.sessionToken {
		t.Fatal("session cookie missing from UI response")
	}
}

func TestStartRequiresSessionToken(t *testing.T) {
	runCalled := false
	s := NewServer(8899, func(_ context.Context, _ JobConfig, _ io.Writer) (JobResult, error) {
		runCalled = true
		return JobResult{}, nil
	})
	r := httptest.NewRequest(http.MethodPost, "/api/start", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	s.handleStart(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
	if runCalled {
		t.Fatal("run called without session token")
	}
}

func TestEventsRequiresSessionToken(t *testing.T) {
	s := NewServer(8899, nil)
	r := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()

	s.handleEvents(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSessionTokenAllowsHeaderOrCookie(t *testing.T) {
	s := NewServer(8899, nil)
	headerRequest := httptest.NewRequest(http.MethodPost, "/api/start", nil)
	headerRequest.Header.Set("X-CSRF-Token", s.sessionToken)
	if !s.validSessionToken(headerRequest) {
		t.Fatal("header token rejected")
	}

	cookieRequest := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	cookieRequest.AddCookie(&http.Cookie{Name: "web_session_token", Value: s.sessionToken})
	if !s.validSessionToken(cookieRequest) {
		t.Fatal("cookie token rejected")
	}
}
