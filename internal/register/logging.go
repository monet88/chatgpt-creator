package register

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	diagnosticWriter    io.Writer = os.Stdout
	diagnosticMu        sync.RWMutex
	diagnosticSessionMu sync.Mutex
)

func SetDiagnosticWriter(writer io.Writer) {
	diagnosticMu.Lock()
	defer diagnosticMu.Unlock()
	if writer == nil {
		diagnosticWriter = os.Stdout
		return
	}
	diagnosticWriter = writer
}

func diagnosticPrintf(format string, args ...any) {
	diagnosticMu.RLock()
	writer := diagnosticWriter
	diagnosticMu.RUnlock()
	fmt.Fprintf(writer, format, args...)
}

func WithDiagnosticWriter(writer io.Writer, run func()) {
	diagnosticSessionMu.Lock()
	defer diagnosticSessionMu.Unlock()

	diagnosticMu.RLock()
	previous := diagnosticWriter
	diagnosticMu.RUnlock()

	SetDiagnosticWriter(writer)
	defer SetDiagnosticWriter(previous)

	run()
}
