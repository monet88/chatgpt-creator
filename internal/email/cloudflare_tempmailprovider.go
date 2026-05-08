package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CloudflareTempMailProvider polls the Cloudflare temp-mail Worker API
// to retrieve OTP codes sent to a @<domain> address managed by the Worker.
type CloudflareTempMailProvider struct {
	baseURL    string // e.g. "https://mail.monet.uno"
	httpClient *http.Client
	pollInterval time.Duration
}

// NewCloudflareTempMailProvider creates a provider pointing at the given Worker base URL.
func NewCloudflareTempMailProvider(baseURL string) *CloudflareTempMailProvider {
	return &CloudflareTempMailProvider{
		baseURL:      strings.TrimRight(baseURL, "/"),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		pollInterval: 5 * time.Second,
	}
}

type cfOTPResponse struct {
	Success bool `json:"success"`
	Data    struct {
		OTP        string `json:"otp"`
		ReceivedAt string `json:"receivedAt"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GetOTP polls the Worker OTP endpoint until a fresh code is received or timeout expires.
// Only OTPs with receivedAt after (now - otpFreshnessBuf) are accepted to reject stale mailbox entries.
func (c *CloudflareTempMailProvider) GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error) {
	parts := strings.SplitN(emailAddr, "@", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("cloudflare-tempmailprovider: invalid email address: %q", emailAddr)
	}
	user, domain := parts[0], parts[1]
	otpURL := fmt.Sprintf("%s/api/v1/email/%s/%s/otp", c.baseURL, domain, user)

	// Accept OTPs received no earlier than 60s before polling starts to tolerate
	// auto-sent OTPs that arrive slightly before GetOTP is called while rejecting
	// stale entries from prior registration attempts on the same mailbox.
	freshSince := time.Now().Add(-60 * time.Second)
	deadline := time.Now().Add(timeout)
	for {
		code, err := c.fetchFreshOTP(ctx, otpURL, freshSince)
		if err == nil && code != "" {
			return code, nil
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("cloudflare-tempmailprovider: OTP timeout after %v for %s", timeout, emailAddr)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(c.pollInterval):
		}
	}
}

// fetchFreshOTP fetches the latest OTP and returns it only if receivedAt is after freshSince.
func (c *CloudflareTempMailProvider) fetchFreshOTP(ctx context.Context, otpURL string, freshSince time.Time) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, otpURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudflare-tempmailprovider: status %d", resp.StatusCode)
	}

	var result cfOTPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("cloudflare-tempmailprovider: decode error: %w", err)
	}

	if !result.Success {
		msg := "unknown error"
		if result.Error != nil {
			msg = result.Error.Message
		}
		return "", fmt.Errorf("cloudflare-tempmailprovider: API error: %s", msg)
	}

	if result.Data.OTP == "" {
		return "", nil
	}

	if result.Data.ReceivedAt != "" {
		receivedAt, parseErr := time.Parse(time.RFC3339Nano, result.Data.ReceivedAt)
		if parseErr == nil && receivedAt.Before(freshSince) {
			return "", nil
		}
	}

	return result.Data.OTP, nil
}

// Close is a no-op.
func (c *CloudflareTempMailProvider) Close() error { return nil }

type cfGenerateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Email  string `json:"email"`
		User   string `json:"user"`
		Domain string `json:"domain"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// CreateEmail calls the Worker API to generate and register a new mailbox, returning the full email address.
func (c *CloudflareTempMailProvider) CreateEmail(domain string) (string, error) {
	type generateRequest struct {
		Domain string `json:"domain,omitempty"`
	}
	payload, _ := json.Marshal(generateRequest{Domain: domain})
	url := c.baseURL + "/api/v1/email/generate"

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("cloudflare-tempmailprovider: generate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cloudflare-tempmailprovider: generate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cloudflare-tempmailprovider: generate status %d", resp.StatusCode)
	}

	var result cfGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("cloudflare-tempmailprovider: generate decode: %w", err)
	}
	if !result.Success {
		msg := "unknown error"
		if result.Error != nil {
			msg = result.Error.Message
		}
		return "", fmt.Errorf("cloudflare-tempmailprovider: generate API error: %s", msg)
	}
	return result.Data.Email, nil
}

// CreateCloudflareTempEmail creates a new mailbox via the Worker API and returns the email address.
// This satisfies the batchDependencies.createTempEmail signature.
func CreateCloudflareTempEmail(baseURL string) func(domain string) (string, error) {
	p := NewCloudflareTempMailProvider(baseURL)
	return p.CreateEmail
}
