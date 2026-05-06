package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/verssache/chatgpt-creator/internal/register"
)

type batchCall struct {
	total           int
	outputFile      string
	maxWorkers      int
	proxy           string
	defaultPassword string
	defaultDomain   string
}

func executeCommandForTest(t *testing.T, args []string, stdin string) (int, string, string) {
	t.Helper()

	in := strings.NewReader(stdin)
	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd := newRegisterCommand(in, &out, &errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err == nil {
		return 0, out.String(), errOut.String()
	}

	var ee *exitError
	if errors.As(err, &ee) {
		errOut.WriteString(ee.Error())
		errOut.WriteString("\n")
		return ee.code, out.String(), errOut.String()
	}

	errOut.WriteString(err.Error())
	errOut.WriteString("\n")
	return exitCodeRuntime, out.String(), errOut.String()
}

func TestCommand_NonInteractiveParsing(t *testing.T) {
	var called bool
	var captured batchCall
	prevRunBatch := runBatchWithProviders
	runBatchWithProviders = func(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts register.BatchOptions, providers register.ProviderOptions) register.BatchResult {
		called = true
		captured = batchCall{
			total:           totalAccounts,
			outputFile:      outputFile,
			maxWorkers:      maxWorkers,
			proxy:           proxy,
			defaultPassword: defaultPassword,
			defaultDomain:   defaultDomain,
		}
		return register.BatchResult{Target: totalAccounts, Success: int64(totalAccounts), Attempts: int64(totalAccounts), Failures: 0, StopReason: register.StopReasonTargetReached}
	}
	t.Cleanup(func() { runBatchWithProviders = prevRunBatch })

	configPath := filepath.Join(t.TempDir(), "config.json")
	content := []byte(`{"proxy":"http://config:8080","output_file":"config-out.txt","default_password":"longpassword12","default_domain":"config.example"}`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	exitCode, stdout, stderr := executeCommandForTest(t, []string{
		"--config", configPath,
		"--total", "10",
		"--workers", "3",
		"--proxy", "http://flag:8080",
		"--output", "flag-out.txt",
		"--password", "flagpassword12",
		"--domain", "flag.example",
		"--json",
	}, "")

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %q", exitCode, stderr)
	}
	if !called {
		t.Fatal("runBatch was not called")
	}
	if captured.total != 10 || captured.maxWorkers != 3 {
		t.Fatalf("captured total/workers = %d/%d", captured.total, captured.maxWorkers)
	}
	if captured.proxy != "http://flag:8080" {
		t.Fatalf("proxy = %q", captured.proxy)
	}
	if captured.outputFile != "flag-out.txt" {
		t.Fatalf("outputFile = %q", captured.outputFile)
	}
	if captured.defaultPassword != "flagpassword12" {
		t.Fatalf("defaultPassword = %q", captured.defaultPassword)
	}
	if captured.defaultDomain != "flag.example" {
		t.Fatalf("defaultDomain = %q", captured.defaultDomain)
	}
	if !strings.Contains(stdout, `"target":10`) {
		t.Fatalf("stdout = %q, want json with target", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, "flagpassword12") || strings.Contains(stderr, "flagpassword12") {
		t.Fatalf("password leaked in output: stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestCommand_MissingTotalReturnsValidationError(t *testing.T) {
	exitCode, _, stderr := executeCommandForTest(t, []string{"--workers", "3"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "--total must be greater than 0") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_InvalidWorkersReturnsValidationError(t *testing.T) {
	exitCode, _, stderr := executeCommandForTest(t, []string{"--total", "1", "--workers", "0"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "--workers must be greater than 0") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_FlagOverridesEnvProxy(t *testing.T) {
	prevProxy := os.Getenv("PROXY")
	t.Cleanup(func() { _ = os.Setenv("PROXY", prevProxy) })
	if err := os.Setenv("PROXY", "http://env:8080"); err != nil {
		t.Fatalf("Setenv() error = %v", err)
	}

	var capturedProxy string
	prevRunBatch := runBatchWithProviders
	runBatchWithProviders = func(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts register.BatchOptions, providers register.ProviderOptions) register.BatchResult {
		capturedProxy = proxy
		return register.BatchResult{Target: totalAccounts, Success: int64(totalAccounts), Attempts: int64(totalAccounts), StopReason: register.StopReasonTargetReached}
	}
	t.Cleanup(func() { runBatchWithProviders = prevRunBatch })

	exitCode, _, stderr := executeCommandForTest(t, []string{"--total", "1", "--workers", "1", "--proxy", "http://flag:8080"}, "")
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %q", exitCode, stderr)
	}
	if capturedProxy != "http://flag:8080" {
		t.Fatalf("proxy = %q", capturedProxy)
	}
}

func TestCommand_InteractiveFallbackUsesStdin(t *testing.T) {
	var called bool
	prevRunBatchForCLI := runBatchForCLI
	runBatchForCLI = func(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts register.BatchOptions) register.BatchResult {
		called = true
		return register.BatchResult{Target: totalAccounts, Success: int64(totalAccounts), Attempts: int64(totalAccounts), StopReason: register.StopReasonTargetReached}
	}
	t.Cleanup(func() { runBatchForCLI = prevRunBatchForCLI })

	configPath := filepath.Join(t.TempDir(), "missing.json")
	// Stdin: proxy, total, workers, password, domain, pacing
	input := strings.Join([]string{"", "1", "", "supersecurepass", "", ""}, "\n") + "\n"
	exitCode, stdout, stderr := executeCommandForTest(t, []string{"--config", configPath}, input)

	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %q", exitCode, stderr)
	}
	if !called {
		t.Fatal("runBatch was not called")
	}
	if !strings.Contains(stdout, "Total accounts to register:") {
		t.Fatalf("stdout = %q", stdout)
	}
	if strings.Contains(stdout, "supersecurepass") {
		t.Fatalf("password leaked in interactive stdout: %q", stdout)
	}
	if !strings.Contains(stdout, "Password:       [redacted]") {
		t.Fatalf("stdout missing redacted password line: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestCommand_CodexFlagFailsClosed(t *testing.T) {
	exitCode, _, stderr := executeCommandForTest(t, []string{"--total", "1", "--workers", "1", "--codex"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "codex extraction is not supported in safe mode") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_CodexConfigFailsClosed(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	content := []byte(`{"codex_enabled":true}`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	exitCode, _, stderr := executeCommandForTest(t, []string{"--config", configPath, "--total", "1", "--workers", "1"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "codex extraction is not supported in safe mode") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_ViOTPFlagFailsClosed(t *testing.T) {
	exitCode, _, stderr := executeCommandForTest(t, []string{"--total", "1", "--workers", "1", "--viotp-token", "token"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "viotp phone challenge automation is not supported in safe mode") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_ViOTPConfigFailsClosed(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	content := []byte(`{"viotp_token":"token-from-config"}`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	exitCode, _, stderr := executeCommandForTest(t, []string{"--config", configPath, "--total", "1", "--workers", "1"}, "")
	if exitCode != exitCodeValidation {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
	}
	if !strings.Contains(stderr, "viotp phone challenge automation is not supported in safe mode") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCommand_ActionableFlagsSkipInteractiveFallback(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{name: "proxy cooldown", args: []string{"--proxy-cooldown", "10"}},
		{name: "imap port", args: []string{"--imap-port", "993"}},
		{name: "imap user", args: []string{"--imap-user", "user@example.com"}},
		{name: "codex output", args: []string{"--codex-output", "unused.json"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exitCode, stdout, stderr := executeCommandForTest(t, tc.args, "")
			if exitCode != exitCodeValidation {
				t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeValidation)
			}
			if !strings.Contains(stderr, "--total must be greater than 0") {
				t.Fatalf("stderr = %q", stderr)
			}
			if strings.Contains(stdout, "Total accounts to register:") {
				t.Fatalf("unexpected interactive prompt in stdout: %q", stdout)
			}
		})
	}
}

func TestCommand_InteractiveFallbackBeforeIMAPInit(t *testing.T) {
	var called bool
	prevRunBatchForCLI := runBatchForCLI
	runBatchForCLI = func(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts register.BatchOptions) register.BatchResult {
		called = true
		return register.BatchResult{Target: totalAccounts, Success: int64(totalAccounts), Attempts: int64(totalAccounts), StopReason: register.StopReasonTargetReached}
	}
	t.Cleanup(func() { runBatchForCLI = prevRunBatchForCLI })

	configPath := filepath.Join(t.TempDir(), "config.json")
	content := []byte(`{"imap_host":"localhost","imap_user":"user","imap_password":"pass","imap_port":65535,"imap_use_tls":false}`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	input := strings.Join([]string{"", "1", "", "supersecurepass", "", ""}, "\n") + "\n"
	exitCode, stdout, stderr := executeCommandForTest(t, []string{"--config", configPath}, input)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %q", exitCode, stderr)
	}
	if !called {
		t.Fatal("runBatch was not called")
	}
	if !strings.Contains(stdout, "Total accounts to register:") {
		t.Fatalf("stdout = %q", stdout)
	}
	if strings.Contains(stderr, "IMAP connection failed") {
		t.Fatalf("stderr should not include IMAP init error: %q", stderr)
	}
}
