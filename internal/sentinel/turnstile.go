package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/monet88/chatgpt-creator/internal/camofox"
)

const (
	sentinelFrameURL     = "https://sentinel.openai.com/sentinel/20260124ceb8/frame.html"
	turnstileSolveTimeout = 30 * time.Second
	turnstilePollInterval = 500 * time.Millisecond
)

// extractTurnstileTokenJS polls the page for a solved Turnstile token using
// multiple extraction strategies since the frame may set it different ways.
const extractTurnstileTokenJS = `(function() {
	var el = document.querySelector('[name="cf-turnstile-response"]');
	if (el && el.value) return el.value;
	try {
		if (typeof turnstile !== 'undefined') {
			var containers = document.querySelectorAll('.cf-turnstile,[data-cf-turnstile]');
			for (var i = 0; i < containers.length; i++) {
				var t = turnstile.getResponse(containers[i]);
				if (t) return t;
			}
			var t2 = turnstile.getResponse();
			if (t2) return t2;
		}
	} catch(_) {}
	if (window.__sentinel_turnstile_token) return window.__sentinel_turnstile_token;
	if (window.turnstileToken) return window.turnstileToken;
	return '';
})()`

// SolveTurnstile opens the sentinel frame in camofox with the dx challenge
// parameter and polls until Turnstile auto-solves. Returns the solved token.
func SolveTurnstile(ctx context.Context, camofoxURL, proxy, deviceID, sentinelCToken, dx string) (string, error) {
	userID := "sentinel-" + deviceID
	if len(deviceID) > 8 {
		userID = "sentinel-" + deviceID[:8]
	}

	client := camofox.NewClient(userID, camofox.WithBaseURL(camofoxURL))

	frameURL := sentinelFrameURL
	params := url.Values{}
	if sentinelCToken != "" {
		params.Set("c", sentinelCToken)
	}
	if dx != "" {
		params.Set("dx", dx)
	}
	if encoded := params.Encode(); encoded != "" {
		frameURL += "?" + encoded
	}

	tabID, err := client.OpenTabWithOptions(ctx, frameURL, "sentinel", camofox.OpenTabOptions{
		ProxyURL: proxy,
	})
	if err != nil {
		return "", fmt.Errorf("turnstile: open sentinel frame: %w", err)
	}
	defer client.CloseTab(context.Background(), tabID) //nolint:errcheck

	deadline := time.Now().Add(turnstileSolveTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		time.Sleep(turnstilePollInterval)

		raw, evalErr := client.Evaluate(ctx, tabID, extractTurnstileTokenJS)
		if evalErr != nil {
			continue
		}

		var token string
		if jsonErr := json.Unmarshal(raw, &token); jsonErr == nil && token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("turnstile: solve timeout after %s", turnstileSolveTimeout)
}
