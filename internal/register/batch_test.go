package register

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{name: "seconds only", in: 45 * time.Second, want: "45s"},
		{name: "minutes and seconds", in: 2*time.Minute + 5*time.Second, want: "2m 5s"},
		{name: "hours minutes seconds", in: time.Hour + 2*time.Minute + 3*time.Second, want: "1h 2m 3s"},
		{name: "zero", in: 0, want: "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.in)
			if got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
