package register

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monet88/chatgpt-creator/internal/codex"
)

// panelEntry is the per-account JSON record compatible with playful-proxy-api-panel.
type panelEntry struct {
	TokenVersion     int64             `json:"_token_version"`
	AccessToken      string            `json:"access_token"`
	AccountID        string            `json:"account_id"`
	ChatGPTAccountID string            `json:"chatgpt_account_id"`
	ChatGPTUserID    string            `json:"chatgpt_user_id"`
	Disabled         bool              `json:"disabled"`
	Email            string            `json:"email"`
	Expired          string            `json:"expired"`
	ExpiresAt        string            `json:"expires_at"`
	IDToken          string            `json:"id_token"`
	LastRefresh      string            `json:"last_refresh"`
	ModelMapping     map[string]string `json:"model_mapping"`
	OrganizationID   string            `json:"organization_id"`
	PlanType         string            `json:"plan_type"`
	RefreshToken     string            `json:"refresh_token"`
	Type             string            `json:"type"`
}

// idTokenClaims holds the fields extracted from the OpenAI id_token JWT payload.
type idTokenClaims struct {
	Email string `json:"email"`
	Auth  struct {
		ChatGPTAccountID string `json:"chatgpt_account_id"`
		ChatGPTUserID    string `json:"chatgpt_user_id"`
		ChatGPTPlanType  string `json:"chatgpt_plan_type"`
		Organizations    []struct {
			ID string `json:"id"`
		} `json:"organizations"`
	} `json:"https://api.openai.com/auth"`
}

// parseIDTokenClaims decodes JWT payload without signature verification.
// Safe here because the token arrives directly from OpenAI's auth server.
func parseIDTokenClaims(idToken string) (*idTokenClaims, error) {
	parts := strings.SplitN(idToken, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("panel: invalid id_token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("panel: decode id_token payload: %w", err)
	}
	var claims idTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("panel: unmarshal id_token claims: %w", err)
	}
	return &claims, nil
}

type modelsResponse struct {
	Models []struct {
		Slug string `json:"slug"`
	} `json:"models"`
}

// fetchPanelModelMapping calls /backend-api/models with the access token and returns
// a slug→slug mapping. Returns an empty map on any error (non-fatal).
func fetchPanelModelMapping(ctx context.Context, accessToken string) map[string]string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://chatgpt.com/backend-api/models", nil)
	if err != nil {
		return map[string]string{}
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]string{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]string{}
	}

	var result modelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]string{}
	}

	mapping := make(map[string]string, len(result.Models))
	for _, m := range result.Models {
		if m.Slug != "" {
			mapping[m.Slug] = m.Slug
		}
	}
	return mapping
}

// buildPanelEntry constructs a panelEntry from a TokenResult.
// If fetchModels is true it calls /backend-api/models; failures are silently ignored.
func buildPanelEntry(ctx context.Context, email string, tokens *codex.TokenResult, fetchModels bool) *panelEntry {
	now := time.Now()
	expiresAt := now.Add(time.Duration(tokens.ExpiresIn) * time.Second)
	// Refresh tokens are valid for ~7 days; use that as the panel "expired" deadline.
	expired := now.Add(7 * 24 * time.Hour)

	entry := &panelEntry{
		TokenVersion: now.UnixMilli(),
		AccessToken:  tokens.AccessToken,
		Disabled:     false,
		Email:        email,
		Expired:      expired.Format(time.RFC3339),
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		IDToken:      tokens.IDToken,
		LastRefresh:  now.Format(time.RFC3339),
		ModelMapping: map[string]string{},
		PlanType:     "free",
		RefreshToken: tokens.RefreshToken,
		Type:         "codex",
	}

	if tokens.IDToken != "" {
		if claims, err := parseIDTokenClaims(tokens.IDToken); err == nil {
			entry.AccountID = claims.Auth.ChatGPTAccountID
			entry.ChatGPTAccountID = claims.Auth.ChatGPTAccountID
			entry.ChatGPTUserID = claims.Auth.ChatGPTUserID
			if claims.Auth.ChatGPTPlanType != "" {
				entry.PlanType = claims.Auth.ChatGPTPlanType
			}
			if len(claims.Auth.Organizations) > 0 {
				entry.OrganizationID = claims.Auth.Organizations[0].ID
			}
		}
	}

	if fetchModels {
		entry.ModelMapping = fetchPanelModelMapping(ctx, tokens.AccessToken)
	}

	return entry
}

// writePanelFile writes a panelEntry to outputDir/codex-{email}-{plan}.json atomically.
func writePanelFile(outputDir string, entry *panelEntry) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("panel: create output dir: %w", err)
	}

	filename := fmt.Sprintf("codex-%s-%s.json", entry.Email, entry.PlanType)
	path := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("panel: marshal entry: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("panel: write file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("panel: rename file: %w", err)
	}
	return nil
}
