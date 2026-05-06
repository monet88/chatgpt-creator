package email

import (
	"context"
	"time"
)

// OTPProvider abstracts email-based OTP retrieval.
type OTPProvider interface {
	// GetOTP retrieves an OTP code sent to the given email address.
	// It polls until the OTP is found or the timeout is reached.
	GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error)
	// Close releases any resources held by the provider.
	Close() error
}

// GeneratorEmailProvider wraps the existing HTML-scraping email provider
// behind the OTPProvider interface for backward compatibility.
type GeneratorEmailProvider struct{}

// GetOTP polls generator.email for a verification code.
func (g *GeneratorEmailProvider) GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error) {
	maxRetries := int(timeout / (3 * time.Second))
	if maxRetries < 1 {
		maxRetries = 1
	}
	return GetVerificationCodeWithContext(ctx, emailAddr, maxRetries, 3*time.Second)
}

// Close is a no-op for the generator email provider.
func (g *GeneratorEmailProvider) Close() error { return nil }
