package register

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monet88/chatgpt-creator/internal/proxy"
)

type fakeFlowRunner struct {
	runErr error
}

func (f fakeFlowRunner) RunRegisterWithContext(ctx context.Context, emailAddr, password, name, birthdate string) error {
	return f.runErr
}

func baseDeps(t *testing.T, runnerErr error, writeFn func(outputFile, emailAddr, password string) error) batchDependencies {
	t.Helper()
	return batchDependencies{
		newClient: func(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error) {
			return fakeFlowRunner{runErr: runnerErr}, nil
		},
		createTempEmail: func(defaultDomain string) (string, error) {
			return "alice@example.com", nil
		},
		generatePassword: func() string {
			return "generated-password"
		},
		randomName: func() (string, string) {
			return "Alice", "Doe"
		},
		randomBirthdate: func() string {
			return "1999-01-01"
		},
		writeCredential: writeFn,
		resolveProxy: func(ctx context.Context, fallback string) (string, error) {
			return fallback, nil
		},
		reportProxy: func(proxyURL string, success bool) {},
	}
}

func TestRegisterOne_SuccessWritesCredential(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "results.txt")
	deps := baseDeps(t, nil, appendCredential)

	var printMu sync.Mutex
	var fileMu sync.Mutex

	success, emailAddr, runErr := registerOne(context.Background(), 1, "1/1", "", outputPath, "", "example.com", &printMu, &fileMu, deps)
	if !success {
		t.Fatalf("success = false, err = %v", runErr)
	}
	if emailAddr != "alice@example.com" {
		t.Fatalf("emailAddr = %q, want %q", emailAddr, "alice@example.com")
	}
	if runErr != nil {
		t.Fatalf("runErr = %v, want nil", runErr)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "alice@example.com|generated-password\n" {
		t.Fatalf("output content = %q", string(content))
	}
}

func TestRegisterOne_FailuresDoNotWriteCredential(t *testing.T) {
	failureCases := []struct {
		name   string
		runErr error
	}{
		{name: "unsupported email", runErr: errors.New("unsupported_email: domain blocked")},
		{name: "otp timeout", runErr: errors.New("failed to get verification code after 20 retries")},
		{name: "challenge failed", runErr: errors.New("challenge failed")},
		{name: "upstream changed", runErr: errors.New("authorize url not found")},
	}

	for _, tc := range failureCases {
		t.Run(tc.name, func(t *testing.T) {
			writeCalled := false
			deps := baseDeps(t, tc.runErr, func(outputFile, emailAddr, password string) error {
				writeCalled = true
				return nil
			})

			var printMu sync.Mutex
			var fileMu sync.Mutex

			success, _, runErr := registerOne(context.Background(), 1, "1/1", "", filepath.Join(t.TempDir(), "results.txt"), "", "example.com", &printMu, &fileMu, deps)
			if success {
				t.Fatal("success = true, want false")
			}
			if runErr == nil || !strings.Contains(runErr.Error(), tc.runErr.Error()) {
				t.Fatalf("runErr = %v, want contains %q", runErr, tc.runErr.Error())
			}
			if writeCalled {
				t.Fatal("writeCredential called on failure")
			}
		})
	}
}

func TestRegisterOne_WriteFailureReturnsError(t *testing.T) {
	deps := baseDeps(t, nil, func(outputFile, emailAddr, password string) error {
		return errors.New("failed to write to output file")
	})

	var printMu sync.Mutex
	var fileMu sync.Mutex

	success, _, runErr := registerOne(context.Background(), 1, "1/1", "", filepath.Join(t.TempDir(), "results.txt"), "", "example.com", &printMu, &fileMu, deps)
	if success {
		t.Fatal("success = true, want false")
	}
	if runErr == nil || !strings.Contains(runErr.Error(), "failed to write to output file") {
		t.Fatalf("runErr = %v", runErr)
	}
}

func TestRunBatchWithOptions_ProxyStatsSnapshot(t *testing.T) {
	statsSnapshot := map[string]proxy.ProxyStats{
		"http://proxy-1:8080": {
			URL:      "http://proxy-1:8080",
			Success:  1,
			Failures: 0,
		},
	}
	deps := baseDeps(t, nil, appendCredential)
	deps.resolveProxy = func(ctx context.Context, fallback string) (string, error) {
		return "http://proxy-1:8080", nil
	}
	deps.proxyStats = func() map[string]proxy.ProxyStats {
		result := make(map[string]proxy.ProxyStats, len(statsSnapshot))
		for key, value := range statsSnapshot {
			result[key] = value
		}
		return result
	}
	deps.reportProxy = func(proxyURL string, success bool) {
		entry := statsSnapshot[proxyURL]
		entry.URL = proxyURL
		if success {
			entry.Success++
		} else {
			entry.Failures++
			entry.CoolUntil = time.Now().Add(5 * time.Minute)
		}
		statsSnapshot[proxyURL] = entry
	}

	result := RunBatchWithOptions(context.Background(), 1, filepath.Join(t.TempDir(), "results.txt"), 1, "", "", "", deps, BatchOptions{MaxAttempts: 4, MaxConsecutiveFailures: 1, PerAccountTimeout: 30 * time.Second, RetryBaseDelay: 0, PacingProfile: PacingNone})
	if len(result.ProxyStats) == 0 {
		t.Fatal("expected proxy stats in result")
	}
	entry, ok := result.ProxyStats["http://proxy-1:8080"]
	if !ok {
		t.Fatal("missing proxy stats for proxy-1")
	}
	if entry.Success < 2 {
		t.Fatalf("success count = %d, want >= 2", entry.Success)
	}
}
