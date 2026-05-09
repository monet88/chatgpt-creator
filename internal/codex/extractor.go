package codex

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	callbackTimeout = 30 * time.Second
)

// TokenEntry is the panel-compatible JSON record written to the output file.
type TokenEntry struct {
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	CreatedAt    string `json:"created_at"`
}

// Extractor orchestrates the full post-registration Codex token extraction flow.
type Extractor struct {
	cfg                 SSOConfig
	outputFile          string
	experimentalEnabled bool
	fileMu              sync.Mutex
}

// NewExtractor creates an Extractor with the default SSO config and output path.
func NewExtractor(outputFile string) *Extractor {
	return &Extractor{
		cfg:                 DefaultSSOConfig(),
		outputFile:          outputFile,
		experimentalEnabled: false,
	}
}

// Extract performs the full PKCE + SSO flow.
// It navigates to the authorize URL via a plain HTTP redirect (not a browser), starts
// a localhost callback server, exchanges the code, and appends the token to the output file.
// Returns a non-nil error if any step fails; callers should treat this as non-fatal.
func (e *Extractor) Extract(ctx context.Context, emailAddr string) (*TokenResult, error) {
	if !e.experimentalEnabled {
		return nil, fmt.Errorf("codex: extraction is not supported in safe mode")
	}

	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("codex: PKCE generation failed: %w", err)
	}

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("codex: state generation failed: %w", err)
	}

	// Bind ephemeral port for callback, then override RedirectURI to match.
	redirectURI, waitFn, listenErr := StartCallbackInterceptor(ctx, state, callbackTimeout)
	if listenErr != nil {
		return nil, fmt.Errorf("codex: %w", listenErr)
	}

	cfg := e.cfg
	cfg.RedirectURI = redirectURI
	authorizeURL := BuildAuthorizeURL(cfg, pkce, state)

	// Wait for callback in background.
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		code, waitErr := waitFn()
		if waitErr != nil {
			errCh <- waitErr
			return
		}
		codeCh <- code
	}()

	// Navigate to authorize URL to trigger OAuth redirect.
	// This plain http.Client call is best-effort only and does not reuse registration session cookies.
	go func() {
		client := &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Follow all redirects — the final one hits our callback server.
				return nil
			},
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, authorizeURL, nil)
		if err != nil {
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()

	// Wait for callback or timeout.
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, fmt.Errorf("codex: callback failed: %w", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	tokens, err := ExchangeCode(ctx, cfg, code, pkce.Verifier)
	if err != nil {
		return nil, fmt.Errorf("codex: token exchange failed: %w", err)
	}

	if err := e.appendToken(emailAddr, tokens); err != nil {
		// Log only; don't fail the whole extraction if the file write fails.
		return tokens, fmt.Errorf("codex: token write failed: %w", err)
	}

	return tokens, nil
}

// appendToken appends a token entry to the JSON output file (creates if absent).
func (e *Extractor) appendToken(emailAddr string, tokens *TokenResult) error {
	e.fileMu.Lock()
	defer e.fileMu.Unlock()

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(tokens.ExpiresIn) * time.Second)
	entry := TokenEntry{
		Email:        emailAddr,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		CreatedAt:    now.Format(time.RFC3339),
	}

	// Read existing entries if the file exists.
	var entries []TokenEntry
	if data, err := os.ReadFile(e.outputFile); err == nil {
		_ = json.Unmarshal(data, &entries)
	}

	entries = append(entries, entry)

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename.
	tmp := e.outputFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, e.outputFile)
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
