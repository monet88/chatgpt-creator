package register

import (
	"strings"
	"testing"
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
