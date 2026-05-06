package register

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/verssache/chatgpt-creator/internal/email"
	"github.com/verssache/chatgpt-creator/internal/util"
)

type flowRunner interface {
	RunRegisterWithContext(ctx context.Context, emailAddr, password, name, birthdate string) error
}

type batchDependencies struct {
	newClient       func(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (flowRunner, error)
	createTempEmail func(defaultDomain string) (string, error)
	generatePassword func() string
	randomName      func() (string, string)
	randomBirthdate func() string
	writeCredential func(outputFile, emailAddr, password string) error
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
	}
}

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
