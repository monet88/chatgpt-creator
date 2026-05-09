package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAuthBaseURL = "https://auth.openai.com"
	defaultClientID    = "app_EMoamEEZ73f0CkXaXp7hrann"
	defaultRedirectURI = "http://127.0.0.1:1455/auth/callback"
	defaultScope       = "openid email profile offline_access"
)

// TokenResult holds the tokens obtained from the OAuth flow.
type TokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// SSOConfig configures the SSO token extraction flow.
type SSOConfig struct {
	AuthBaseURL string
	ClientID    string
	RedirectURI string
	Scope       string
}

// DefaultSSOConfig returns the default Codex SSO configuration.
func DefaultSSOConfig() SSOConfig {
	return SSOConfig{
		AuthBaseURL: defaultAuthBaseURL,
		ClientID:    defaultClientID,
		RedirectURI: defaultRedirectURI,
		Scope:       defaultScope,
	}
}

// BuildAuthorizeURL constructs the OAuth2 /oauth/authorize URL with PKCE parameters.
func BuildAuthorizeURL(cfg SSOConfig, pkce *PKCE, state string) string {
	params := url.Values{
		"response_type":              {"code"},
		"client_id":                  {cfg.ClientID},
		"redirect_uri":               {cfg.RedirectURI},
		"scope":                      {cfg.Scope},
		"state":                      {state},
		"code_challenge":             {pkce.Challenge},
		"code_challenge_method":      {pkce.Method},
		"id_token_add_organizations": {"true"},
		"codex_cli_simplified_flow":  {"true"},
	}
	return cfg.AuthBaseURL + "/oauth/authorize?" + params.Encode()
}

// InterceptCallback starts a temporary HTTP server to capture the OAuth callback.
// Returns the authorization code when received, or error on timeout/cancellation.
func InterceptCallback(ctx context.Context, listenAddr string, expectedState string, timeout time.Duration) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	callbackHandler := func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("sso: state mismatch: got %q, want %q", state, expectedState)
			return
		}

		errParam := r.URL.Query().Get("error")
		if errParam != "" {
			desc := r.URL.Query().Get("error_description")
			http.Error(w, "OAuth error: "+desc, http.StatusBadRequest)
			errCh <- fmt.Errorf("sso: OAuth error: %s (%s)", errParam, desc)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- fmt.Errorf("sso: callback missing authorization code")
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authorization successful</h1><p>You can close this window.</p></body></html>")
		codeCh <- code
	}
	mux.HandleFunc("/auth/callback", callbackHandler)
	mux.HandleFunc("/callback", callbackHandler)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return "", fmt.Errorf("sso: failed to listen on %s: %w", listenAddr, err)
	}

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("sso: server error: %w", err)
		}
	}()

	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(shutCtx)
	}()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("sso: callback timeout after %v", timeout)
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// StartCallbackInterceptor binds a local listener on an OS-assigned ephemeral port,
// returning the redirect URI and a wait function. Each concurrent worker gets its own
// port, eliminating fixed-port conflicts with other processes.
func StartCallbackInterceptor(ctx context.Context, expectedState string, timeout time.Duration) (redirectURI string, wait func() (string, error), err error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, fmt.Errorf("sso: bind callback listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI = fmt.Sprintf("http://127.0.0.1:%d/auth/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("sso: state mismatch: got %q, want %q", state, expectedState):
			default:
			}
			return
		}

		errParam := r.URL.Query().Get("error")
		if errParam != "" {
			desc := r.URL.Query().Get("error_description")
			http.Error(w, "OAuth error: "+desc, http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("sso: OAuth error: %s (%s)", errParam, desc):
			default:
			}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			select {
			case errCh <- fmt.Errorf("sso: callback missing authorization code"):
			default:
			}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h1>Authorization successful</h1><p>You can close this window.</p></body></html>")
		select {
		case codeCh <- code:
		default:
		}
	})

	server := &http.Server{Handler: mux}
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			select {
			case errCh <- fmt.Errorf("sso: server error: %w", serveErr):
			default:
			}
		}
	}()

	wait = func() (string, error) {
		defer func() {
			shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = server.Shutdown(shutCtx)
		}()
		select {
		case code := <-codeCh:
			return code, nil
		case cbErr := <-errCh:
			return "", cbErr
		case <-time.After(timeout):
			return "", fmt.Errorf("sso: callback timeout after %v", timeout)
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return redirectURI, wait, nil
}

// ExchangeCode exchanges the authorization code for tokens using the PKCE verifier.
func ExchangeCode(ctx context.Context, cfg SSOConfig, code string, verifier string) (*TokenResult, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {cfg.ClientID},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURI},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.AuthBaseURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("sso: failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sso: token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sso: failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sso: token exchange failed (status=%d)", resp.StatusCode)
	}

	var result TokenResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("sso: failed to parse token response: %w", err)
	}

	return &result, nil
}
