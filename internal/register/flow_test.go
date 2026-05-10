package register

import (
	"strings"
	"testing"
)

func TestAuthorizePathRouting(t *testing.T) {
	tests := []struct {
		name          string
		finalPath     string
		wantRegister  bool
		wantOTPOnly   bool
		wantAboutYou  bool
	}{
		{name: "password screen", finalPath: "/u/create-account/password", wantRegister: true},
		{name: "email verification", finalPath: "/u/email-verification", wantOTPOnly: true},
		{name: "email otp", finalPath: "/u/email-otp", wantOTPOnly: true},
		{name: "about you", finalPath: "/u/about-you", wantAboutYou: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegister := strings.Contains(tt.finalPath, "create-account/password")
			gotOTPOnly := strings.Contains(tt.finalPath, "email-verification") || strings.Contains(tt.finalPath, "email-otp")
			gotAboutYou := strings.Contains(tt.finalPath, "about-you")

			if gotRegister != tt.wantRegister {
				t.Errorf("register path %q: got %v want %v", tt.finalPath, gotRegister, tt.wantRegister)
			}
			if gotOTPOnly != tt.wantOTPOnly {
				t.Errorf("otp-only path %q: got %v want %v", tt.finalPath, gotOTPOnly, tt.wantOTPOnly)
			}
			if gotAboutYou != tt.wantAboutYou {
				t.Errorf("about-you path %q: got %v want %v", tt.finalPath, gotAboutYou, tt.wantAboutYou)
			}
		})
	}
}
