package register

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/monet88/chatgpt-creator/internal/email"
	"github.com/monet88/chatgpt-creator/internal/phone"
	proxypkg "github.com/monet88/chatgpt-creator/internal/proxy"
	"github.com/monet88/chatgpt-creator/internal/util"
)

type BatchOptions struct {
	MaxAttempts            int
	MaxConsecutiveFailures int
	PerAccountTimeout      time.Duration
	RetryBaseDelay         time.Duration
	PacingProfile          PacingProfile
}

func defaultBatchOptions(totalAccounts int) BatchOptions {
	maxAttempts := totalAccounts * 4
	if maxAttempts < totalAccounts {
		maxAttempts = totalAccounts
	}
	return BatchOptions{
		MaxAttempts:            maxAttempts,
		MaxConsecutiveFailures: totalAccounts,
		PerAccountTimeout:      90 * time.Second,
		RetryBaseDelay:         500 * time.Millisecond,
		PacingProfile:          PacingHuman,
	}
}

// DefaultBatchOptionsForCLI returns default batch options for CLI usage.
func DefaultBatchOptionsForCLI(totalAccounts int) BatchOptions {
	return defaultBatchOptions(totalAccounts)
}

// registerOne handles a single account registration.
func registerOne(ctx context.Context, workerID int, tag string, proxy, outputFile, defaultPassword, defaultDomain string, printMu, fileMu *sync.Mutex, deps batchDependencies) (bool, string, error) {
	resolvedProxy, err := deps.resolveProxy(ctx, proxy)
	if err != nil {
		return false, "", WrapFailure("resolve_proxy", 0, err)
	}

	client, err := newClientWithDeps(deps, resolvedProxy, tag, workerID, printMu, fileMu)
	if err != nil {
		deps.reportProxy(resolvedProxy, false)
		return false, "", WrapFailure("new_client", 0, err)
	}

	emailAddr, mailboxURL, err := deps.createTempEmail(defaultDomain)
	if err != nil {
		return false, "", WrapFailure("create_temp_email", 0, err)
	}

	password := defaultPassword
	if password == "" {
		password = deps.generatePassword()
	}

	firstName, lastName, parsed := util.ParseNameFromEmail(emailAddr)
	if !parsed {
		firstName, lastName = deps.randomName()
	}
	birthdate := deps.randomBirthdate()

	err = client.RunRegisterWithContext(ctx, emailAddr, password, firstName+" "+lastName, birthdate)
	if err != nil {
		deps.reportProxy(resolvedProxy, false)
		return false, emailAddr, WrapFailure("run_register", 0, err)
	}

	deps.reportProxy(resolvedProxy, true)

	fileMu.Lock()
	err = deps.writeCredential(outputFile, emailAddr, password, mailboxURL)
	fileMu.Unlock()
	if err != nil {
		return false, emailAddr, NewFailure(FailureOutputWrite, "write_credential", 0, err)
	}

	return true, emailAddr, nil
}

func RunBatchForCLI(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts BatchOptions) BatchResult {
	return RunBatchWithOptions(
		ctx,
		totalAccounts,
		outputFile,
		maxWorkers,
		proxy,
		defaultPassword,
		defaultDomain,
		defaultBatchDependencies(),
		opts,
	)
}

// ProviderOptions holds optional provider overrides for batch registration.
// Zero values use defaults (generator.email OTP, no proxy pool, no codex).
type ProviderOptions struct {
	// ProxyPool overrides the single proxy with a rotating pool.
	// When set, resolveProxy delegates to pool.Next and reportProxy delegates to pool.Report.
	ProxyPool interface {
		Next(ctx context.Context) (string, error)
		Report(proxyURL string, success bool)
		Stats() map[string]proxypkg.ProxyStats
	}
	// OTPProvider overrides the default generator.email OTP retrieval.
	OTPProvider email.OTPProvider
	// PhoneProvider handles SMS-based phone verification challenges.
	PhoneProvider phone.PhoneProvider
	// ViOTPServiceID is the ViOTP service ID used for OpenAI SMS verification.
	ViOTPServiceID int
	// CodexEnabled enables post-registration Codex OAuth token extraction.
	CodexEnabled bool
	// CodexOutput is the JSON file path where Codex tokens are appended.
	CodexOutput string
	// PanelOutputDir is the directory where per-account panel JSON files are written.
	// Files are named codex-{email}-{plan}.json. No-op when empty.
	PanelOutputDir string
	// CreateTempEmail overrides the default temp email creation function.
	// When set, it replaces the generator.email-based mailbox creation.
	CreateTempEmail func(domain string) (emailAddr, mailboxURL string, err error)
}

// RunBatchForCLIWithProviders is like RunBatchForCLI but accepts optional provider overrides.
func RunBatchForCLIWithProviders(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, opts BatchOptions, providers ProviderOptions) BatchResult {
	deps := defaultBatchDependencies()

	if providers.OTPProvider != nil {
		deps.otpProvider = providers.OTPProvider
	}
	if providers.CreateTempEmail != nil {
		deps.createTempEmail = providers.CreateTempEmail
	}
	if providers.PhoneProvider != nil {
		deps.phoneProvider = providers.PhoneProvider
		deps.viOTPServiceID = providers.ViOTPServiceID
	}
	if providers.CodexEnabled {
		deps.codexEnabled = true
		deps.codexOutput = providers.CodexOutput
		deps.panelOutputDir = providers.PanelOutputDir
	}
	if providers.ProxyPool != nil {
		pool := providers.ProxyPool
		deps.resolveProxy = func(ctx context.Context, fallback string) (string, error) {
			return pool.Next(ctx)
		}
		deps.reportProxy = func(proxyURL string, success bool) {
			pool.Report(proxyURL, success)
		}
		deps.proxyStats = pool.Stats
	}

	return RunBatchWithOptions(
		ctx,
		totalAccounts,
		outputFile,
		maxWorkers,
		proxy,
		defaultPassword,
		defaultDomain,
		deps,
		opts,
	)
}

// RunBatch runs concurrent registration tasks with retry until target success count is reached.
func RunBatch(totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string) {
	RunBatchForCLI(context.Background(), totalAccounts, outputFile, maxWorkers, proxy, defaultPassword, defaultDomain, defaultBatchOptions(totalAccounts))
}

func RunBatchWithDependencies(totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, deps batchDependencies) {
	RunBatchWithOptions(
		context.Background(),
		totalAccounts,
		outputFile,
		maxWorkers,
		proxy,
		defaultPassword,
		defaultDomain,
		deps,
		defaultBatchOptions(totalAccounts),
	)
}

func RunBatchWithOptions(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string, deps batchDependencies, options BatchOptions) BatchResult {
	var printMu sync.Mutex
	var fileMu sync.Mutex
	var failureSummaryMu sync.Mutex

	var remaining int64 = int64(totalAccounts)
	var successCount int64
	var failureCount int64
	var attemptNum int64
	var consecutiveFailures int64
	var stopFlag int32

	failureSummary := map[FailureKind]int64{}
	stopReason := StopReasonTargetReached

	setStop := func(reason StopReason) {
		if atomic.CompareAndSwapInt32(&stopFlag, 0, 1) {
			stopReason = reason
		}
	}

	startTime := time.Now()

	var wg sync.WaitGroup

	for w := 1; w <= maxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				if atomic.LoadInt32(&stopFlag) == 1 {
					return
				}

				select {
				case <-ctx.Done():
					setStop(StopReasonContextCancelled)
					return
				default:
				}

				if options.MaxAttempts > 0 && int(atomic.LoadInt64(&attemptNum)) >= options.MaxAttempts {
					setStop(StopReasonMaxAttemptsReached)
					return
				}

				if atomic.AddInt64(&remaining, -1) < 0 {
					atomic.AddInt64(&remaining, 1)
					return
				}

				attempt := atomic.AddInt64(&attemptNum, 1)
				if options.MaxAttempts > 0 && int(attempt) > options.MaxAttempts {
					atomic.AddInt64(&remaining, 1)
					setStop(StopReasonMaxAttemptsReached)
					return
				}

				tag := fmt.Sprintf("%d/%d", attempt, totalAccounts)

				accountCtx := ctx
				cancel := func() {}
				if options.PerAccountTimeout > 0 {
					accountCtx, cancel = context.WithTimeout(ctx, options.PerAccountTimeout)
				}

				success, emailAddr, runErr := registerOne(accountCtx, workerID, tag, proxy, outputFile, defaultPassword, defaultDomain, &printMu, &fileMu, deps)
				cancel()
				if success {
					atomic.StoreInt64(&consecutiveFailures, 0)
					atomic.AddInt64(&successCount, 1)
					ts := time.Now().Format("15:04:05")
					printMu.Lock()
					diagnosticPrintf("[%s] [W%d] ✓ SUCCESS: %s\n", ts, workerID, safeLogMessage(emailAddr))
					printMu.Unlock()

					// Pacing delay between registrations
					if delay := pacingDelay(options.PacingProfile); delay > 0 {
						if err := waitWithContext(ctx, delay); err != nil {
							setStop(StopReasonContextCancelled)
							return
						}
					}
					continue
				}

				atomic.AddInt64(&failureCount, 1)
				currentConsecutive := atomic.AddInt64(&consecutiveFailures, 1)
				ts := time.Now().Format("15:04:05")
				kind := ClassifyFailureKind(runErr)

				failureSummaryMu.Lock()
				failureSummary[kind]++
				failureSummaryMu.Unlock()

				if IsFailureKind(runErr, FailureUnsupportedEmail) || strings.Contains(strings.ToLower(runErr.Error()), "unsupported_email") {
					parts := strings.Split(emailAddr, "@")
					if len(parts) == 2 {
						domain := parts[1]
						email.AddBlacklistDomain(domain)
						printMu.Lock()
						diagnosticPrintf("[%s] [W%d] ⚠ Blacklisted domain: %s\n", ts, workerID, safeLogMessage(domain))
						printMu.Unlock()
					}
				}

				printMu.Lock()
				diagnosticPrintf("[%s] [W%d] ✗ FAILURE: %s | %s\n", ts, workerID, safeLogMessage(emailAddr), safeLogMessage(runErr.Error()))
				printMu.Unlock()

				if options.MaxConsecutiveFailures > 0 && int(currentConsecutive) >= options.MaxConsecutiveFailures {
					setStop(StopReasonFailureThresholdHit)
					return
				}

				if options.MaxAttempts > 0 && int(atomic.LoadInt64(&attemptNum)) >= options.MaxAttempts {
					setStop(StopReasonMaxAttemptsReached)
					return
				}

				atomic.AddInt64(&remaining, 1)

				// Pacing delay between registrations (even on failure)
				if pDelay := pacingDelay(options.PacingProfile); pDelay > 0 {
					if err := waitWithContext(ctx, pDelay); err != nil {
						setStop(StopReasonContextCancelled)
						return
					}
				}

				// Backoff delay on failure (stacks with pacing)
				delay := backoffDelay(options.RetryBaseDelay, int(currentConsecutive))
				if err := waitWithContext(ctx, delay); err != nil {
					setStop(StopReasonContextCancelled)
					return
				}
			}
		}(w)
	}

	wg.Wait()

	if ctx.Err() != nil {
		setStop(StopReasonContextCancelled)
	}

	elapsed := time.Since(startTime)
	var proxyStats map[string]proxypkg.ProxyStats
	if deps.proxyStats != nil {
		snapshot := deps.proxyStats()
		if len(snapshot) > 0 {
			proxyStats = make(map[string]proxypkg.ProxyStats, len(snapshot))
			for key, value := range snapshot {
				proxyStats[key] = value
			}
		}
	}

	result := BatchResult{
		Target:         totalAccounts,
		Success:        successCount,
		Attempts:       attemptNum,
		Failures:       failureCount,
		Elapsed:        formatDuration(elapsed),
		StopReason:     stopReason,
		OutputFile:     outputFile,
		FailureSummary: failureSummary,
		ProxyStats:     proxyStats,
	}

	diagnosticPrintf("\n--- Batch Registration Summary ---\n")
	diagnosticPrintf("Target:    %d\n", result.Target)
	diagnosticPrintf("Success:   %d\n", result.Success)
	diagnosticPrintf("Attempts:  %d\n", result.Attempts)
	diagnosticPrintf("Failures:  %d\n", result.Failures)
	diagnosticPrintf("Elapsed:   %s\n", result.Elapsed)
	diagnosticPrintf("Stop:      %s\n", result.StopReason)
	diagnosticPrintf("----------------------------------\n")

	return result
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
