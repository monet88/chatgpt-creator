package register

import "testing"

func TestRedactHelpers(t *testing.T) {
	if got := RedactPassword("secret-password"); got != "[redacted]" {
		t.Fatalf("RedactPassword() = %q", got)
	}
	if got := RedactPassword(""); got != "(random)" {
		t.Fatalf("RedactPassword(empty) = %q", got)
	}

	proxy := "http://user:pass@example.com:8080"
	gotProxy := RedactProxy(proxy)
	if gotProxy == proxy {
		t.Fatalf("proxy not redacted: %q", gotProxy)
	}

	line := "token=abc123\nnext"
	safe := safeLogMessage(line)
	if safe == line {
		t.Fatalf("safeLogMessage did not sanitize line: %q", safe)
	}
	if safe == "" {
		t.Fatal("safeLogMessage returned empty")
	}
}
