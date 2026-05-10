package camofox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:9377"

type Client struct {
	BaseURL    string
	UserID     string
	HTTPClient *http.Client
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/"); trimmed != "" {
			c.BaseURL = trimmed
		}
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.HTTPClient = httpClient
		}
	}
}

func NewClient(userID string, opts ...Option) *Client {
	client := &Client{
		BaseURL:    defaultBaseURL,
		UserID:     userID,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *Client) CloseTab(ctx context.Context, tabID string) error {
	return c.doJSON(ctx, http.MethodDelete, "/tabs/"+url.PathEscape(tabID), map[string]string{
		"userId": c.UserID,
	}, nil)
}

func (c *Client) GetTabURL(ctx context.Context, tabID string) (string, error) {
	var out struct {
		Tabs []struct {
			ID    string `json:"id"`
			TabID string `json:"tabId"`
			URL   string `json:"url"`
		} `json:"tabs"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/tabs?userId="+url.QueryEscape(c.UserID), nil, &out); err != nil {
		return "", err
	}
	for _, tab := range out.Tabs {
		if tab.ID == tabID || tab.TabID == tabID {
			return tab.URL, nil
		}
	}
	return "", fmt.Errorf("camofox: tab %q not found", tabID)
}

func (c *Client) Wait(ctx context.Context, tabID string, timeout time.Duration) error {
	return c.doJSON(ctx, http.MethodPost, "/tabs/"+url.PathEscape(tabID)+"/wait", map[string]any{
		"userId":         c.UserID,
		"timeout":        timeout.Milliseconds(),
		"waitForNetwork": true,
	}, nil)
}

func (c *Client) Snapshot(ctx context.Context, tabID string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/tabs/"+url.PathEscape(tabID)+"/snapshot?userId="+url.QueryEscape(c.UserID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("camofox: GET /tabs/%s/snapshot failed (%d): %s", tabID, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.RawMessage(body), nil
}

func (c *Client) TypeText(ctx context.Context, tabID, selector, text string) error {
	return c.doJSON(ctx, http.MethodPost, "/tabs/"+url.PathEscape(tabID)+"/type", map[string]string{
		"userId":   c.UserID,
		"selector": selector,
		"text":     text,
	}, nil)
}

func (c *Client) Click(ctx context.Context, tabID, selector string) error {
	return c.doJSON(ctx, http.MethodPost, "/tabs/"+url.PathEscape(tabID)+"/click", map[string]string{
		"userId":   c.UserID,
		"selector": selector,
	}, nil)
}

func (c *Client) Navigate(ctx context.Context, tabID, navigateURL string) error {
	return c.doJSON(ctx, http.MethodPost, "/tabs/"+url.PathEscape(tabID)+"/navigate", map[string]any{
		"userId": c.UserID,
		"url":    navigateURL,
	}, nil)
}

func (c *Client) Evaluate(ctx context.Context, tabID, expression string) (json.RawMessage, error) {
	var out struct {
		Result json.RawMessage `json:"result"`
		Value  json.RawMessage `json:"value"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/tabs/"+url.PathEscape(tabID)+"/evaluate", map[string]any{
		"userId":     c.UserID,
		"expression": expression,
	}, &out); err != nil {
		return nil, err
	}
	if len(out.Result) > 0 {
		return out.Result, nil
	}
	return out.Value, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body, out any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, payload)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("camofox: %s %s failed (%d): %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
