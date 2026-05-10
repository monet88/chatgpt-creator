package mfa

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type stubOTPProvider struct {
	codes []string
	index int
}

func (s *stubOTPProvider) GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error) {
	if s.index >= len(s.codes) {
		return "", errors.New("no more OTPs")
	}
	code := s.codes[s.index]
	s.index++
	return code, nil
}

func (s *stubOTPProvider) Close() error { return nil }

func TestNextLoginOTP_SkipsPreviousCode(t *testing.T) {
	provider := &stubOTPProvider{codes: []string{"111111", "111111", "222222"}}

	got, err := nextLoginOTP(context.Background(), "alice@example.com", provider, 10*time.Second, "111111")
	if err != nil {
		t.Fatalf("nextLoginOTP() error = %v", err)
	}
	if got != "222222" {
		t.Fatalf("nextLoginOTP() = %q, want %q", got, "222222")
	}
}

type errOTPProvider struct{}

func (e *errOTPProvider) GetOTP(ctx context.Context, emailAddr string, timeout time.Duration) (string, error) {
	return "", errors.New("provider unavailable")
}

func (e *errOTPProvider) Close() error { return nil }

func TestNextLoginOTP_ReturnsProviderError(t *testing.T) {
	_, err := nextLoginOTP(context.Background(), "alice@example.com", &errOTPProvider{}, time.Second, "111111")
	if err == nil || err.Error() == "" {
		t.Fatal("nextLoginOTP() error = nil, want provider failure")
	}
	if want := "provider unavailable"; !strings.Contains(err.Error(), want) {
		t.Fatalf("nextLoginOTP() error = %v, want contains %q", err, want)
	}
}
