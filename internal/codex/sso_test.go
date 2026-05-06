package codex

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() error = %v", err)
	}

	if pkce.Method != "S256" {
		t.Fatalf("Method = %q, want S256", pkce.Method)
	}
	if len(pkce.Verifier) < 43 {
		t.Fatalf("Verifier too short: %d", len(pkce.Verifier))
	}
	if pkce.Challenge == "" {
		t.Fatal("Challenge is empty")
	}

	// Verify challenge = base64url(sha256(verifier))
	h := sha256.Sum256([]byte(pkce.Verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h[:])
	if pkce.Challenge != expectedChallenge {
		t.Fatalf("Challenge mismatch: got %q, want %q", pkce.Challenge, expectedChallenge)
	}
}

func TestGeneratePKCE_Uniqueness(t *testing.T) {
	p1, _ := GeneratePKCE()
	p2, _ := GeneratePKCE()
	if p1.Verifier == p2.Verifier {
		t.Fatal("two PKCE pairs should have different verifiers")
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	cfg := DefaultSSOConfig()
	pkce := &PKCE{
		Verifier:  "test-verifier",
		Challenge: "test-challenge",
		Method:    "S256",
	}

	authURL := BuildAuthorizeURL(cfg, pkce, "test-state")

	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.Host != "auth0.openai.com" {
		t.Fatalf("Host = %q", parsed.Host)
	}
	if parsed.Path != "/authorize" {
		t.Fatalf("Path = %q", parsed.Path)
	}

	params := parsed.Query()
	if params.Get("response_type") != "code" {
		t.Fatalf("response_type = %q", params.Get("response_type"))
	}
	if params.Get("code_challenge") != "test-challenge" {
		t.Fatalf("code_challenge = %q", params.Get("code_challenge"))
	}
	if params.Get("state") != "test-state" {
		t.Fatalf("state = %q", params.Get("state"))
	}
}

func TestInterceptCallback_Success(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	codeCh := make(chan string)
	errCh := make(chan error)

	go func() {
		code, err := InterceptCallback(context.Background(), addr, "test-state", 5*time.Second)
		if err != nil {
			errCh <- err
		} else {
			codeCh <- code
		}
	}()

	// Wait briefly for server to start
	time.Sleep(100 * time.Millisecond)

	// Simulate callback
	callbackURL := fmt.Sprintf("http://%s/callback?code=test-auth-code&state=test-state", addr)
	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("callback request error = %v", err)
	}
	resp.Body.Close()

	select {
	case code := <-codeCh:
		if code != "test-auth-code" {
			t.Fatalf("code = %q, want test-auth-code", code)
		}
	case err := <-errCh:
		t.Fatalf("InterceptCallback() error = %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for callback result")
	}
}

func TestInterceptCallback_StateMismatch(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := InterceptCallback(context.Background(), addr, "correct-state", 2*time.Second)
		errCh <- err
	}()

	time.Sleep(100 * time.Millisecond)

	callbackURL := fmt.Sprintf("http://%s/callback?code=test&state=wrong-state", addr)
	resp, _ := http.Get(callbackURL)
	if resp != nil {
		resp.Body.Close()
	}

	select {
	case err := <-errCh:
		if err == nil || !strings.Contains(err.Error(), "state mismatch") {
			t.Fatalf("expected state mismatch error, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	}
}

func TestInterceptCallback_Timeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	_, err = InterceptCallback(context.Background(), addr, "state", 100*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestExchangeCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		params, _ := url.ParseQuery(string(body))
		if params.Get("grant_type") != "authorization_code" {
			t.Fatalf("grant_type = %q", params.Get("grant_type"))
		}
		if params.Get("code_verifier") != "test-verifier" {
			t.Fatalf("code_verifier = %q", params.Get("code_verifier"))
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"at-123","refresh_token":"rt-456","token_type":"Bearer","expires_in":3600}`)
	}))
	defer server.Close()

	cfg := DefaultSSOConfig()
	cfg.AuthBaseURL = server.URL

	result, err := ExchangeCode(context.Background(), cfg, "test-code", "test-verifier")
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if result.AccessToken != "at-123" {
		t.Fatalf("AccessToken = %q", result.AccessToken)
	}
	if result.RefreshToken != "rt-456" {
		t.Fatalf("RefreshToken = %q", result.RefreshToken)
	}
	if result.ExpiresIn != 3600 {
		t.Fatalf("ExpiresIn = %d", result.ExpiresIn)
	}
}

func TestExchangeCode_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"code expired"}`)
	}))
	defer server.Close()

	cfg := DefaultSSOConfig()
	cfg.AuthBaseURL = server.URL

	_, err := ExchangeCode(context.Background(), cfg, "expired-code", "verifier")
	if err == nil {
		t.Fatal("expected error for expired code")
	}
}
