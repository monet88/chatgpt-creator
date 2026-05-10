package camofox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type OpenTabOptions struct {
	ProxyURL string
}

func (c *Client) OpenTab(ctx context.Context, openURL, sessionKey string) (string, error) {
	return c.OpenTabWithOptions(ctx, openURL, sessionKey, OpenTabOptions{})
}

func (c *Client) OpenTabWithOptions(ctx context.Context, openURL, sessionKey string, opts OpenTabOptions) (string, error) {
	var out struct {
		ID    string `json:"id"`
		TabID string `json:"tabId"`
	}
	if sessionKey == "" {
		sessionKey = "default"
	}
	req := map[string]any{
		"userId":     c.UserID,
		"sessionKey": sessionKey,
		"url":        openURL,
	}
	if proxy, err := parseProxy(opts.ProxyURL); err != nil {
		return "", err
	} else if proxy != nil {
		req["proxy"] = proxy
		req["geoMode"] = "proxy-locked"
	}
	if err := c.doJSON(ctx, "POST", "/tabs", req, &out); err != nil {
		return "", err
	}
	if out.TabID != "" {
		return out.TabID, nil
	}
	if out.ID != "" {
		return out.ID, nil
	}
	return "", fmt.Errorf("camofox: open tab returned no tab id")
}

func parseProxy(proxyURL string) (map[string]any, error) {
	if proxyURL == "" {
		return nil, nil
	}
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("camofox: invalid proxy URL: %w", err)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || parsed.Hostname() == "" || port <= 0 {
		return nil, fmt.Errorf("camofox: proxy URL must include host and port")
	}
	proxy := map[string]any{
		"host": parsed.Hostname(),
		"port": port,
	}
	if parsed.User != nil {
		if user := parsed.User.Username(); user != "" {
			proxy["username"] = user
		}
		if password, ok := parsed.User.Password(); ok && password != "" {
			proxy["password"] = password
		}
	}
	return proxy, nil
}
