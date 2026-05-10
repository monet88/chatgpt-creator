package register

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/monet88/chatgpt-creator/internal/camofox"
	"github.com/monet88/chatgpt-creator/internal/codex"
	"github.com/monet88/chatgpt-creator/internal/mfa"
	"github.com/monet88/chatgpt-creator/internal/util"
)

const codexBrowserTimeout = 90 * time.Second

// codex1455Mu serializes all browser-based Codex extractions so only one
// worker binds port 1455 at a time.
var codex1455Mu sync.Mutex

// extractCodexViaBrowser runs the Codex PKCE OAuth flow using a camofox browser
// session. It forces a fresh login (prompt=login) and handles the full auth
// flow: email → password → email OTP (if needed) → TOTP (if MFA enabled) →
// consent → callback interception on localhost:1455.
func (c *Client) extractCodexViaBrowser(ctx context.Context, emailAddr string) error {
	pkce, err := codex.GeneratePKCE()
	if err != nil {
		return fmt.Errorf("codex: PKCE generation failed: %w", err)
	}
	state := util.GenerateUUID()

	cfg := codex.DefaultSSOConfig()
	cfg.Prompt = "login"
	authorizeURL := codex.BuildAuthorizeURL(cfg, pkce, state)

	codex1455Mu.Lock()
	defer codex1455Mu.Unlock()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, cbErr := codex.InterceptCallback(ctx, "127.0.0.1:1455", state, codexBrowserTimeout)
		if cbErr != nil {
			errCh <- cbErr
		} else {
			codeCh <- code
		}
	}()

	camofoxUserID := strings.NewReplacer("@", "-", ".", "-").Replace(emailAddr)
	browser := camofox.NewClient(camofoxUserID, camofox.WithBaseURL(c.camofoxURL))

	tabID, openErr := browser.OpenTabWithOptions(ctx, authorizeURL, "codex-login", camofox.OpenTabOptions{ProxyURL: c.proxy})
	if openErr != nil {
		return fmt.Errorf("codex: open browser tab: %w", openErr)
	}
	defer browser.CloseTab(context.Background(), tabID)

	pollDeadline := time.Now().Add(codexBrowserTimeout)
	var lastURL string

	for time.Now().Before(pollDeadline) {
		select {
		case code := <-codeCh:
			return c.finalizeCodex(ctx, cfg, code, pkce.Verifier, emailAddr)
		case cbErr := <-errCh:
			return fmt.Errorf("codex: %w", cbErr)
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tabURL, urlErr := browser.GetTabURL(ctx, tabID)
		if urlErr != nil || tabURL == "" {
			select {
			case <-time.After(1500 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}

		c.print("Codex browser: " + tabURL)

		if strings.Contains(tabURL, "auth.openai.com/error") {
			return fmt.Errorf("codex: auth error page — run with --mfa or check camofox session")
		}

		// Fast-path: code already in callback URL.
		if parsed, pErr := url.Parse(tabURL); pErr == nil {
			if code := parsed.Query().Get("code"); code != "" {
				return c.finalizeCodex(ctx, cfg, code, pkce.Verifier, emailAddr)
			}
		}

		// Only interact when URL changes.
		if tabURL == lastURL {
			select {
			case <-time.After(1500 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		lastURL = tabURL

		switch {
		case strings.Contains(tabURL, "/log-in/password") || strings.Contains(tabURL, "log-in/identifier"):
			_ = browser.TypeText(ctx, tabID, `input[type="password"]`, c.accountPassword)
			_ = browser.Click(ctx, tabID, `button[type="submit"]`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)

		case strings.Contains(tabURL, "email-verification") || strings.Contains(tabURL, "/email-otp") || strings.Contains(tabURL, "/otp"):
			otp, otpErr := c.otpProvider.GetOTP(ctx, emailAddr, 5*time.Minute)
			if otpErr != nil {
				return fmt.Errorf("codex: OTP fetch failed: %w", otpErr)
			}
			_ = browser.TypeText(ctx, tabID, `input[autocomplete="one-time-code"]`, otp)
			_ = browser.Click(ctx, tabID, `button[type="submit"]`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)

		case strings.Contains(tabURL, "phone-verification") || strings.Contains(tabURL, "/phone") || strings.Contains(tabURL, "sms-verification"):
			if c.phoneProvider == nil {
				return fmt.Errorf("codex: phone verification required but no phone provider configured (use --viotp-token)")
			}
			rental, rentErr := c.phoneProvider.RentNumber(ctx, c.viOTPServiceID)
			if rentErr != nil {
				return fmt.Errorf("codex: rent phone number failed: %w", rentErr)
			}
			c.print("Codex: phone rented ..." + maskPhone(rental.PhoneNumber))
			// Select Vietnam (+84): try click-based dropdown first, fall back to JS.
			selectVietnam(ctx, browser, tabID)
			_ = browser.TypeText(ctx, tabID, `input[type="tel"], input[name="phone"]`, rental.PhoneNumber)
			_ = browser.Click(ctx, tabID, `button[type="submit"]`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)
			smsOTP, smsErr := c.phoneProvider.WaitForOTP(ctx, rental.RequestID, phoneOTPTimeout)
			if smsErr != nil {
				return fmt.Errorf("codex: SMS OTP timeout: %w", smsErr)
			}
			_ = browser.TypeText(ctx, tabID, `input[autocomplete="one-time-code"], input[type="tel"]`, smsOTP)
			_ = browser.Click(ctx, tabID, `button[type="submit"]`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)

		case c.totpSecret != "" && (strings.Contains(tabURL, "totp") || strings.Contains(tabURL, "/mfa")):
			totpCode, totpErr := mfa.GenerateTOTP(c.totpSecret)
			if totpErr == nil {
				_ = browser.TypeText(ctx, tabID, `input[autocomplete="one-time-code"]`, totpCode)
				_ = browser.Click(ctx, tabID, `button[type="submit"]`)
				_ = browser.Wait(ctx, tabID, 5*time.Second)
			}

		case strings.Contains(tabURL, "consent"):
			_ = browser.Click(ctx, tabID, `button:has-text("Continue")`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)

		case strings.Contains(tabURL, "auth.openai.com") && !strings.Contains(tabURL, "/oauth/authorize"):
			// Login page (log-in, sign-up, etc.) — type email and continue.
			_ = browser.TypeText(ctx, tabID, `input[type="email"]`, emailAddr)
			_ = browser.Click(ctx, tabID, `button[type="submit"]`)
			_ = browser.Wait(ctx, tabID, 5*time.Second)
		}

		select {
		case <-time.After(1500 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Final wait for the callback that may arrive just after poll loop exits.
	select {
	case code := <-codeCh:
		return c.finalizeCodex(ctx, cfg, code, pkce.Verifier, emailAddr)
	case cbErr := <-errCh:
		return fmt.Errorf("codex: %w", cbErr)
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("codex: timed out waiting for OAuth callback")
	}
}

// selectVietnam selects Vietnam (+84) in the phone country-code picker.
// Tries click-based dropdown interaction first; falls back to JS assignment
// for standard <select> elements.
func selectVietnam(ctx context.Context, browser *camofox.Client, tabID string) {
	// 1. Try clicking the country selector trigger (flag button / code button).
	_ = browser.Click(ctx, tabID,
		`button[aria-haspopup="listbox"], button[aria-label*="country" i], button[aria-label*="Country" i], `+
			`[data-testid*="country"], button:has-text("+1"), button:has-text("US"), button:has-text("United States")`)
	_ = browser.Wait(ctx, tabID, 1*time.Second)

	// 2. Click the Vietnam option inside the opened dropdown.
	_ = browser.Click(ctx, tabID,
		`li:has-text("+84"), li:has-text("Vietnam"), option[value="VN"], `+
			`[role="option"]:has-text("Vietnam"), [role="option"]:has-text("+84")`)
	_ = browser.Wait(ctx, tabID, 500*time.Millisecond)

	// 3. JS fallback for plain <select> elements.
	_, _ = browser.Evaluate(ctx, tabID, `
		const s = document.querySelector('select[name="countryCode"],select[name="country"],select[name="phone_country"]');
		if (s) { s.value = 'VN'; s.dispatchEvent(new Event('change', {bubbles:true})); }
	`)
}

func (c *Client) finalizeCodex(ctx context.Context, cfg codex.SSOConfig, code, verifier, emailAddr string) error {
	tokens, err := codex.ExchangeCode(ctx, cfg, code, verifier)
	if err != nil {
		return fmt.Errorf("codex: token exchange failed: %w", err)
	}
	return c.writeCodexArtifacts(ctx, emailAddr, tokens)
}
