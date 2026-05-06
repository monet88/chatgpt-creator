package codex

import (
	"context"
	"strings"
	"testing"
)

func TestExtractor_Extract_FailsClosedInSafeMode(t *testing.T) {
	extractor := NewExtractor("ignored.json")

	result, err := extractor.Extract(context.Background(), "alice@example.com")
	if err == nil {
		t.Fatal("expected safe-mode unsupported error")
	}
	if result != nil {
		t.Fatalf("result = %#v, want nil", result)
	}
	if !strings.Contains(err.Error(), "not supported in safe mode") {
		t.Fatalf("err = %q, want safe-mode unsupported message", err.Error())
	}
}
