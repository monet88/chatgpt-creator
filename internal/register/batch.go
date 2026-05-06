package register

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/verssache/chatgpt-creator/internal/email"
)

type BatchOptions struct {
	MaxAttempts            int
	MaxConsecutiveFailures int
	PerAccountTimeout      time.Duration
	RetryBaseDelay         time.Duration
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
	}
}

// registerOne handles a single account registration.
func registerOne(ctx context.Context, workerID int, tag string, proxy, outputFile, defaultPassword, defaultDomain string, printMu, fileMu *sync.Mutex, deps batchDependencies) (bool, string, error) {
	client, err := deps.newClient(proxy, tag, workerID, printMu, fileMu)
	if err != nil {
		return false, "", WrapFailure("new_client", 0, err)
	}

	emailAddr, err := deps.createTempEmail(defaultDomain)
	if err != nil {
		return false, "", WrapFailure("create_temp_email", 0, err)
	}

	password := defaultPassword
	if password == "" {
		password = deps.generatePassword()
	}

	firstName, lastName := deps.randomName()
	birthdate := deps.randomBirthdate()

	err = client.RunRegisterWithContext(ctx, emailAddr, password, firstName+" "+lastName, birthdate)
	if err != nil {
		return false, emailAddr, WrapFailure("run_register", 0, err)
	}

	fileMu.Lock()
	err = deps.writeCredential(outputFile, emailAddr, password)
	fileMu.Unlock()
	if err != nil {
		return false, emailAddr, NewFailure(FailureOutputWrite, "write_credential", 0, err)
	}

	return true, emailAddr, nil
}

func RunBatchForCLI(ctx context.Context, totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string) BatchResult {
	return RunBatchWithOptions(
		ctx,
		totalAccounts,
		outputFile,
		maxWorkers,
		proxy,
		defaultPassword,
		defaultDomain,
		defaultBatchDependencies(),
		defaultBatchOptions(totalAccounts),
	)
}

// RunBatch runs concurrent registration tasks with retry until target success count is reached.
func RunBatch(totalAccounts int, outputFile string, maxWorkers int, proxy, defaultPassword, defaultDomain string) {
	RunBatchForCLI(context.Background(), totalAccounts, outputFile, maxWorkers, proxy, defaultPassword, defaultDomain)
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
	result := BatchResult{
		Target:         totalAccounts,
		Success:        successCount,
		Attempts:       attemptNum,
		Failures:       failureCount,
		Elapsed:        formatDuration(elapsed),
		StopReason:     stopReason,
		OutputFile:     outputFile,
		FailureSummary: failureSummary,
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
