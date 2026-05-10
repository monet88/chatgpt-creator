package mfa

import (
	"testing"
	"time"
)

func TestGenerateTOTPAt_RFC6238Vectors(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	tests := []struct {
		at   int64
		want string
	}{
		{at: 59, want: "287082"},
		{at: 1111111109, want: "081804"},
		{at: 1111111111, want: "050471"},
		{at: 1234567890, want: "005924"},
		{at: 2000000000, want: "279037"},
		{at: 20000000000, want: "353130"},
	}

	for _, tt := range tests {
		got, err := GenerateTOTPAt(secret, time.Unix(tt.at, 0).UTC())
		if err != nil {
			t.Fatalf("GenerateTOTPAt(%d) error = %v", tt.at, err)
		}
		if got != tt.want {
			t.Fatalf("GenerateTOTPAt(%d) = %q, want %q", tt.at, got, tt.want)
		}
	}
}
