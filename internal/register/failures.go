package register

import (
	"errors"
	"fmt"
	"strings"
)

type FailureKind string

const (
	FailureUnsupportedEmail FailureKind = "unsupported_email"
	FailureOTPTimeout       FailureKind = "otp_timeout"
	FailureChallengeFailed  FailureKind = "challenge_failed"
	FailureRateLimited      FailureKind = "rate_limited"
	FailureUpstreamChanged  FailureKind = "upstream_changed"
	FailureNetwork          FailureKind = "network"
	FailureValidation       FailureKind = "validation"
	FailureOutputWrite      FailureKind = "output_write"
	FailureUnknown          FailureKind = "unknown"
)

type FailureError struct {
	Kind   FailureKind
	Step   string
	Status int
	Cause  error
}

func (e *FailureError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		if e.Status > 0 {
			return fmt.Sprintf("%s (%s, status=%d)", e.Step, e.Kind, e.Status)
		}
		return fmt.Sprintf("%s (%s)", e.Step, e.Kind)
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s (%s, status=%d): %v", e.Step, e.Kind, e.Status, e.Cause)
	}
	return fmt.Sprintf("%s (%s): %v", e.Step, e.Kind, e.Cause)
}

func (e *FailureError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NewFailure(kind FailureKind, step string, status int, cause error) error {
	return &FailureError{Kind: kind, Step: step, Status: status, Cause: cause}
}

func IsFailureKind(err error, kind FailureKind) bool {
	var failureErr *FailureError
	if errors.As(err, &failureErr) {
		return failureErr.Kind == kind
	}
	return false
}

func AsFailure(err error) (*FailureError, bool) {
	var failureErr *FailureError
	if errors.As(err, &failureErr) {
		return failureErr, true
	}
	return nil, false
}

func WrapFailure(step string, status int, err error) error {
	if err == nil {
		return nil
	}
	if failureErr, ok := AsFailure(err); ok {
		// Clone to avoid mutating a shared error pointer across goroutines.
		clone := *failureErr
		if clone.Step == "" {
			clone.Step = step
		}
		if clone.Status == 0 {
			clone.Status = status
		}
		return &clone
	}
	return NewFailure(ClassifyFailureKind(err), step, status, err)
}

func ClassifyFailureKind(err error) FailureKind {
	if err == nil {
		return FailureUnknown
	}
	if IsFailureKind(err, FailureUnsupportedEmail) ||
		IsFailureKind(err, FailureOTPTimeout) ||
		IsFailureKind(err, FailureChallengeFailed) ||
		IsFailureKind(err, FailureRateLimited) ||
		IsFailureKind(err, FailureUpstreamChanged) ||
		IsFailureKind(err, FailureNetwork) ||
		IsFailureKind(err, FailureValidation) ||
		IsFailureKind(err, FailureOutputWrite) {
		failureErr, _ := AsFailure(err)
		return failureErr.Kind
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unsupported_email"):
		return FailureUnsupportedEmail
	case strings.Contains(message, "verification code") && strings.Contains(message, "failed"):
		return FailureOTPTimeout
	case strings.Contains(message, "sentinel") || strings.Contains(message, "challenge"):
		return FailureChallengeFailed
	case strings.Contains(message, "429") || strings.Contains(message, "rate"):
		return FailureRateLimited
	case strings.Contains(message, "authorize url not found") || strings.Contains(message, "unknown jump"):
		return FailureUpstreamChanged
	case strings.Contains(message, "invalid") || strings.Contains(message, "must be"):
		return FailureValidation
	case strings.Contains(message, "output file") || strings.Contains(message, "write to output"):
		return FailureOutputWrite
	case strings.Contains(message, "connection") || strings.Contains(message, "timeout") || strings.Contains(message, "request"):
		return FailureNetwork
	default:
		return FailureUnknown
	}
}
