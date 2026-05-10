package mfa

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/monet88/chatgpt-creator/internal/camofox"
	"github.com/monet88/chatgpt-creator/internal/email"
)

var secretPattern = regexp.MustCompile(`[A-Z2-7]{4}(?: ?[A-Z2-7]{4}){3,}`)

type EnrollResult struct {
	Secret   string
	FactorID string
}

type SetupOpts struct {
	CamofoxBaseURL string
	OTPTimeout     time.Duration
	Password       string
	Proxy          string
}

func SetupTOTP(ctx context.Context, emailAddr string, otpProvider email.OTPProvider, camofoxUserID string, opts SetupOpts) (*EnrollResult, error) {
	if otpProvider == nil {
		return nil, fmt.Errorf("mfa: otp provider is required")
	}
	if strings.TrimSpace(camofoxUserID) == "" {
		return nil, fmt.Errorf("mfa: camofox user id is required")
	}

	browser := camofox.NewClient(camofoxUserID, camofox.WithBaseURL(opts.CamofoxBaseURL))
	const sessionKey = "mfa-setup"
	loginTab, err := browser.OpenTabWithOptions(ctx, "https://chatgpt.com", sessionKey, camofox.OpenTabOptions{ProxyURL: opts.Proxy})
	if err != nil {
		return nil, err
	}
	tabs := []string{loginTab}
	defer closeTabs(ctx, browser, tabs)

	if err := browser.Wait(ctx, loginTab, 8*time.Second); err != nil {
		return nil, err
	}
	loginURL, text, err := pageState(ctx, browser, loginTab)
	if err != nil {
		return nil, err
	}
	if needsLogin(loginURL, text) {
		if err := typeAny(ctx, browser, loginTab, emailAddr, `input[name="email"]`, `input[type="email"]`); err != nil {
			if clickErr := clickAny(ctx, browser, loginTab, `a:has-text("Log in")`, `button:has-text("Log in")`); clickErr != nil {
				return nil, fmt.Errorf("mfa: open login: %w", err)
			}
			if err = browser.Wait(ctx, loginTab, 8*time.Second); err != nil {
				return nil, err
			}
			if err = typeAny(ctx, browser, loginTab, emailAddr, `input[name="email"]`, `input[type="email"]`); err != nil {
				return nil, err
			}
		}
		if err := clickAny(ctx, browser, loginTab, `button[type="submit"]`, `button:has-text("Continue")`); err != nil {
			return nil, err
		}
		if err := browser.Wait(ctx, loginTab, 8*time.Second); err != nil {
			return nil, err
		}
		loginURL, text, err = pageState(ctx, browser, loginTab)
		if err != nil {
			return nil, err
		}
		if strings.Contains(loginURL, "/password") || strings.Contains(strings.ToLower(text), "password") {
			if strings.TrimSpace(opts.Password) == "" {
				return nil, fmt.Errorf("mfa: password challenge requires a password")
			}
			if err := typeAny(ctx, browser, loginTab, opts.Password, `input[type="password"]`); err != nil {
				return nil, err
			}
			if err := clickAny(ctx, browser, loginTab, `button[type="submit"]`, `button:has-text("Continue")`); err != nil {
				return nil, err
			}
			if err := browser.Wait(ctx, loginTab, 8*time.Second); err != nil {
				return nil, err
			}
			loginURL, text, err = pageState(ctx, browser, loginTab)
			if err != nil {
				return nil, err
			}
		}
		if strings.Contains(loginURL, "email-verification") || strings.Contains(loginURL, "/otp") || strings.Contains(strings.ToLower(text), "verification code") {
			if err := submitLoginOTP(ctx, browser, loginTab, emailAddr, otpProvider, opts); err != nil {
				return nil, err
			}
		}
	}

	setupTab, err := browser.OpenTabWithOptions(ctx, "https://chatgpt.com/?action=enable&factor=totp#settings/Security", sessionKey, camofox.OpenTabOptions{ProxyURL: opts.Proxy})
	if err != nil {
		return nil, err
	}
	tabs = append(tabs, setupTab)
	if err := browser.Wait(ctx, setupTab, 8*time.Second); err != nil {
		return nil, err
	}
	clickOptional(ctx, browser, setupTab, `button:has-text("Authenticator app")`)
	clickOptional(ctx, browser, setupTab, `button:has-text("Turn on")`)
	clickOptional(ctx, browser, setupTab, `button:has-text("Continue")`)
	clickOptional(ctx, browser, setupTab, `button:has-text("Set up manually")`, `button:has-text("Enter setup key manually")`, `button:has-text("Can't scan")`)

	_, text, err = pageState(ctx, browser, setupTab)
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(text)
	if strings.Contains(lower, "already enabled") || strings.Contains(lower, "disable two-factor") {
		return nil, fmt.Errorf("mfa: authenticator app already enabled")
	}

	secret := normalizeSecret(secretPattern.FindString(text))
	if secret == "" {
		return nil, fmt.Errorf("mfa: manual setup key not found in snapshot: %s", excerpt(text))
	}

	for attempt := 0; attempt < 3; attempt++ {
		code, err := GenerateTOTP(secret)
		if err != nil {
			return nil, err
		}
		if err := typeAny(ctx, browser, setupTab, code, `input[name="code"], input[autocomplete="one-time-code"]`); err != nil {
			return nil, err
		}
		if err := clickAny(ctx, browser, setupTab, `button:has-text("Verify")`, `button:has-text("Continue")`, `button[type="submit"]`); err != nil {
			return nil, err
		}
		if err := browser.Wait(ctx, setupTab, 8*time.Second); err != nil {
			return nil, err
		}
		_, text, err = pageState(ctx, browser, setupTab)
		if err != nil {
			return nil, err
		}
		lower = strings.ToLower(text)
		if strings.Contains(lower, "recovery code") || strings.Contains(lower, "two-factor authentication is on") || strings.Contains(lower, "disable two-factor") {
			return &EnrollResult{Secret: secret}, nil
		}
		if attempt < 2 {
			if err := sleep(ctx, time.Until(time.Now().Truncate(30*time.Second).Add(30*time.Second))+500*time.Millisecond); err != nil {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("mfa: verification did not reach a success state: %s", excerpt(text))
}

func otpTimeout(opts SetupOpts) time.Duration {
	if opts.OTPTimeout > 0 {
		return opts.OTPTimeout
	}
	return 5 * time.Minute
}
