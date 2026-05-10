package mfa

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monet88/chatgpt-creator/internal/camofox"
	"github.com/monet88/chatgpt-creator/internal/email"
)

func pageState(ctx context.Context, browser *camofox.Client, tabID string) (string, string, error) {
	pageURL, err := browser.GetTabURL(ctx, tabID)
	if err != nil {
		return "", "", err
	}
	snapshot, err := browser.Snapshot(ctx, tabID)
	if err != nil {
		return "", "", err
	}
	return pageURL, flattenSnapshot(snapshot), nil
}

func needsLogin(pageURL, text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(pageURL, "auth.openai.com") || strings.Contains(lower, "log in") || strings.Contains(lower, "continue with email")
}

func typeAny(ctx context.Context, browser *camofox.Client, tabID, value string, selectors ...string) error {
	var lastErr error
	for _, selector := range selectors {
		if err := browser.TypeText(ctx, tabID, selector, value); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func clickAny(ctx context.Context, browser *camofox.Client, tabID string, selectors ...string) error {
	var lastErr error
	for _, selector := range selectors {
		if err := browser.Click(ctx, tabID, selector); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func clickOptional(ctx context.Context, browser *camofox.Client, tabID string, selectors ...string) {
	if err := clickAny(ctx, browser, tabID, selectors...); err == nil {
		_ = browser.Wait(ctx, tabID, 5*time.Second)
	}
}

func waitForURL(ctx context.Context, browser *camofox.Client, tabID string, match func(string) bool, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pageURL, err := browser.GetTabURL(ctx, tabID)
		if err == nil && match(pageURL) {
			return pageURL, nil
		}
		if err := sleep(ctx, 1500*time.Millisecond); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("mfa: wait for URL timed out after %s", timeout)
}

func submitLoginOTP(ctx context.Context, browser *camofox.Client, tabID, emailAddr string, otpProvider email.OTPProvider, opts SetupOpts) error {
	var previous string
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		otp, err := nextLoginOTP(ctx, emailAddr, otpProvider, otpTimeout(opts), previous)
		if err != nil {
			return err
		}
		previous = otp
		if err := typeAny(ctx, browser, tabID, otp, `input[name="code"], input[autocomplete="one-time-code"]`); err != nil {
			return err
		}
		if err := clickAny(ctx, browser, tabID, `button[name="intent"][value="validate"]`, `button[type="submit"]`); err != nil {
			return err
		}
		if _, err := waitForURL(ctx, browser, tabID, func(u string) bool {
			return strings.Contains(u, "chatgpt.com") && !strings.Contains(u, "auth.openai.com")
		}, 20*time.Second); err == nil {
			return nil
		} else {
			lastErr = err
		}
		clickOptional(ctx, browser, tabID, `button:has-text("Resend email")`, `button[name="intent"][value="resend"]`)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("mfa: login OTP validation failed")
}

func nextLoginOTP(ctx context.Context, emailAddr string, otpProvider email.OTPProvider, timeout time.Duration, previous string) (string, error) {
	deadline := time.Now().Add(timeout)
	var lastOTP string
	var lastErr error
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining > 20*time.Second {
			remaining = 20 * time.Second
		}
		otp, err := otpProvider.GetOTP(ctx, emailAddr, remaining)
		if err != nil {
			lastErr = err
		} else if otp != "" {
			lastOTP = otp
			if otp != previous {
				return otp, nil
			}
		}
		if err := sleep(ctx, 3*time.Second); err != nil {
			return "", err
		}
	}
	if lastOTP != "" {
		return "", fmt.Errorf("mfa: OTP provider kept returning the previous code")
	}
	if lastErr != nil {
		return "", fmt.Errorf("mfa: waiting for a fresh login OTP failed: %w", lastErr)
	}
	return "", fmt.Errorf("mfa: timed out waiting for a fresh login OTP")
}

func flattenSnapshot(snapshot json.RawMessage) string {
	var data any
	if err := json.Unmarshal(snapshot, &data); err != nil {
		return string(snapshot)
	}
	var parts []string
	var walk func(any)
	walk = func(value any) {
		switch v := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(v); trimmed != "" {
				parts = append(parts, trimmed)
			}
		case []any:
			for _, item := range v {
				walk(item)
			}
		case map[string]any:
			for _, item := range v {
				walk(item)
			}
		}
	}
	walk(data)
	return strings.Join(parts, " ")
}

func normalizeSecret(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), " ", "")
}

func excerpt(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 200 {
		return value[:200] + "..."
	}
	return value
}

func closeTabs(ctx context.Context, browser *camofox.Client, tabs []string) {
	for i := len(tabs) - 1; i >= 0; i-- {
		_ = browser.CloseTab(ctx, tabs[i])
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
