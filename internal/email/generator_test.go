package email

import (
	"strings"
	"testing"
)

func TestCreateTempEmail_WithDefaultDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
	}{
		{name: "simple domain", domain: "example.com"},
		{name: "subdomain", domain: "mail.example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailAddr, err := CreateTempEmail(tt.domain)
			if err != nil {
				t.Fatalf("CreateTempEmail() error = %v", err)
			}

			parts := strings.Split(emailAddr, "@")
			if len(parts) != 2 {
				t.Fatalf("email format invalid: %q", emailAddr)
			}
			if parts[0] == "" {
				t.Fatalf("local part empty: %q", emailAddr)
			}
			if parts[1] != tt.domain {
				t.Fatalf("domain = %q, want %q", parts[1], tt.domain)
			}
		})
	}
}
