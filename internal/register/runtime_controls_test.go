package register

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countedRunner struct {
	calls *int64
	err   error
}

func (r countedRunner) RunRegisterWithContext(ctx context.Context, emailAddr, password, name, birthdate string) error {
	atomic.AddInt64(r.calls, 1)
	if r.err != nil {
		return r.err
	}
	return nil
}

func newCountedDeps(calls *int64, runErr error) batchDependencies {
	return batchDependencies{
		newClient: func(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error) {
			return countedRunner{calls: calls, err: runErr}, nil
		},
		createTempEmail: func(defaultDomain string) (emailAddr, mailboxURL string, err error) {
			return "alice@example.com", "https://generator.email/example.com/alice", nil
		},
		generatePassword: func() string { return "generated-password" },
		randomName:       func() (string, string) { return "Alice", "Doe" },
		randomBirthdate:  func() string { return "1990-01-01" },
		writeCredential:  func(outputFile, emailAddr, password, mailboxURL string, totpSecret ...string) error { return nil },
		resolveProxy: func(ctx context.Context, fallback string) (string, error) {
			return fallback, nil
		},
		reportProxy: func(proxyURL string, success bool) {},
	}
}

func TestRunBatchWithOptions_StopsAtMaxAttempts(t *testing.T) {
	var calls int64
	deps := newCountedDeps(&calls, errors.New("unsupported_email"))
	options := BatchOptions{
		MaxAttempts:            3,
		MaxConsecutiveFailures: 100,
		PerAccountTimeout:      0,
		RetryBaseDelay:         0,
	}

	RunBatchWithOptions(context.Background(), 1, "results.txt", 1, "", "", "", deps, options)

	if got := atomic.LoadInt64(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestRunBatchWithOptions_StopsAtConsecutiveFailureThreshold(t *testing.T) {
	var calls int64
	deps := newCountedDeps(&calls, errors.New("network failure"))
	options := BatchOptions{
		MaxAttempts:            100,
		MaxConsecutiveFailures: 2,
		PerAccountTimeout:      0,
		RetryBaseDelay:         0,
	}

	RunBatchWithOptions(context.Background(), 1, "results.txt", 1, "", "", "", deps, options)

	if got := atomic.LoadInt64(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestRunBatchWithOptions_StopsOnContextCancel(t *testing.T) {
	var calls int64
	deps := newCountedDeps(&calls, errors.New("network failure"))
	options := BatchOptions{
		MaxAttempts:            100,
		MaxConsecutiveFailures: 100,
		PerAccountTimeout:      0,
		RetryBaseDelay:         0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunBatchWithOptions(ctx, 1, "results.txt", 1, "", "", "", deps, options)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunBatchWithOptions did not stop after context cancellation")
	}
}

func TestFailureClassification(t *testing.T) {
	err := WrapFailure("register", 400, errors.New("unsupported_email"))
	if !IsFailureKind(err, FailureUnsupportedEmail) {
		t.Fatalf("expected FailureUnsupportedEmail, got %v", ClassifyFailureKind(err))
	}

	otpErr := NewFailure(FailureOTPTimeout, "get_otp", 0, errors.New("timeout"))
	if !IsFailureKind(otpErr, FailureOTPTimeout) {
		t.Fatal("expected FailureOTPTimeout")
	}
}

func TestClassifyStatusFailure_PhoneChallenge(t *testing.T) {
	kind := classifyStatusFailure(403, map[string]interface{}{"error": "phone verification challenge required"})
	if kind != FailurePhoneChallenge {
		t.Fatalf("kind = %v, want %v", kind, FailurePhoneChallenge)
	}
}

func TestClassifyFailureKind_PhoneChallenge(t *testing.T) {
	err := NewFailure(FailurePhoneChallenge, "register", 403, errors.New("phone verification challenge required"))
	if got := ClassifyFailureKind(err); got != FailurePhoneChallenge {
		t.Fatalf("got = %v, want %v", got, FailurePhoneChallenge)
	}
}
