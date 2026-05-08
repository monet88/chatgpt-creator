package register

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/monet88/chatgpt-creator/internal/email"
	"github.com/monet88/chatgpt-creator/internal/phone"
	"github.com/monet88/chatgpt-creator/internal/proxy"
	"github.com/monet88/chatgpt-creator/internal/util"
)

type flowRunner interface {
	RunRegisterWithContext(ctx context.Context, emailAddr, password, name, birthdate string) error
}

type batchDependencies struct {
	newClient        func(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error)
	createTempEmail  func(defaultDomain string) (string, error)
	generatePassword func() string
	randomName       func() (string, string)
	randomBirthdate  func() string
	writeCredential  func(outputFile, emailAddr, password string) error
	resolveProxy     func(ctx context.Context, fallback string) (string, error)
	reportProxy      func(proxyURL string, success bool)
	proxyStats       func() map[string]proxy.ProxyStats
	otpProvider      email.OTPProvider
	phoneProvider    phone.PhoneProvider
	viOTPServiceID   int
	codexEnabled    bool
	codexOutput     string
	panelOutputDir  string
}

func defaultBatchDependencies() batchDependencies {
	return batchDependencies{
		newClient: func(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error) {
			return NewClient(proxy, tag, workerID, printMu, fileMu)
		},
		createTempEmail: email.CreateTempEmail,
		generatePassword: func() string {
			return util.GeneratePassword(14)
		},
		randomName:      util.RandomName,
		randomBirthdate: util.RandomBirthdate,
		writeCredential: appendCredential,
		resolveProxy: func(ctx context.Context, fallback string) (string, error) {
			return fallback, nil
		},
		reportProxy: func(proxyURL string, success bool) {},
		proxyStats:  nil,
		otpProvider: &email.GeneratorEmailProvider{},
	}
}

// newClientWithDeps creates a Client and injects dependency overrides from batchDependencies.
func newClientWithDeps(deps batchDependencies, proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error) {
	client, err := deps.newClient(proxy, tag, workerID, printMu, fileMu)
	if err != nil {
		return nil, err
	}
	if c, ok := client.(*Client); ok {
		c.otpProvider    = deps.otpProvider
		c.phoneProvider  = deps.phoneProvider
		c.viOTPServiceID = deps.viOTPServiceID
		c.codexEnabled   = deps.codexEnabled
		c.codexOutput    = deps.codexOutput
		c.panelOutputDir = deps.panelOutputDir
	}
	return client, nil
}

// noopOTPProvider is never nil and always returns an error (used as sentinel).
var _ email.OTPProvider = (*email.GeneratorEmailProvider)(nil)

// defaultOTPTimeout is used when calling otpProvider.GetOTP.
const defaultOTPTimeout = 60 * time.Second

func appendCredential(outputFile, emailAddr, password string) error {
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer f.Close()

	line := fmt.Sprintf("%s|%s\n", emailAddr, password)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to output file: %w", err)
	}

	return nil
}
