package email

import (
	"strings"
	"testing"
)

func TestParseVerificationCodeFromHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "valid otp",
			html: `<div id="email-table"><div class="e7m list-group-item list-group-item-info"><div class="e7m subj_div_45g45gg">Your code is 123456</div></div></div>`,
			want: "123456",
		},
		{
			name: "ignore sentinel otp",
			html: `<div id="email-table"><div class="e7m list-group-item list-group-item-info"><div class="e7m subj_div_45g45gg">Code 177010</div><div class="e7m subj_div_45g45gg">Code 654321</div></div></div>`,
			want: "654321",
		},
		{
			name: "no otp",
			html: `<div id="email-table"><div class="e7m list-group-item list-group-item-info"><div class="e7m subj_div_45g45gg">No numeric code</div></div></div>`,
			want: "",
		},
		{
			name: "malformed html",
			html: `<div id="email-table"><div class="e7m list-group-item list-group-item-info"><div class="e7m subj_div_45g45gg">broken 121212`,
			want: "121212",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVerificationCodeFromHTML(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("parseVerificationCodeFromHTML() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
