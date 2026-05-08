package register

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/monet88/chatgpt-creator/internal/codex"
)

func TestPanelFilenameDistinguishesTruncatedEmails(t *testing.T) {
	prefix := strings.Repeat("a", maxPanelFilenamePartLen)
	first := &panelEntry{Email: prefix + "-first@example.com", PlanType: "plus"}
	second := &panelEntry{Email: prefix + "-second@example.com", PlanType: "plus"}

	if safePanelFilenamePart(first.Email) != safePanelFilenamePart(second.Email) {
		t.Fatal("test setup should produce colliding sanitized filename parts")
	}

	if panelFilename(first) == panelFilename(second) {
		t.Fatalf("panelFilename() collided for distinct emails: %q", panelFilename(first))
	}
}

func TestWriteCodexArtifacts_WritesPanelWhenTokenOutputEmpty(t *testing.T) {
	var printMu sync.Mutex
	var fileMu sync.Mutex
	outputDir := t.TempDir()
	client := &Client{
		printMu:        &printMu,
		fileMu:         &fileMu,
		panelOutputDir: outputDir,
	}

	err := client.writeCodexArtifacts(context.Background(), "alice@example.com", &codex.TokenResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	})
	if err != nil {
		t.Fatalf("writeCodexArtifacts() error = %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(outputDir, "codex-*.json"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("panel files = %d, want 1", len(matches))
	}
}
