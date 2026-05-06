package register

import (
	"testing"
	"time"
)

func TestParsePacingProfile(t *testing.T) {
	tests := []struct {
		input   string
		want    PacingProfile
		wantErr bool
	}{
		{"none", PacingNone, false},
		{"fast", PacingFast, false},
		{"human", PacingHuman, false},
		{"slow", PacingSlow, false},
		{"HUMAN", PacingHuman, false},
		{"  Fast  ", PacingFast, false},
		{"", "", true},
		{"turbo", "", true},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePacingProfile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePacingProfile(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParsePacingProfile(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPacingDelay(t *testing.T) {
	tests := []struct {
		profile PacingProfile
		minDur  time.Duration
		maxDur  time.Duration
	}{
		{PacingNone, 0, 0},
		{PacingFast, 5 * time.Second, 15 * time.Second},
		{PacingHuman, 120 * time.Second, 300 * time.Second},
		{PacingSlow, 300 * time.Second, 600 * time.Second},
	}

	for _, tt := range tests {
		t.Run(string(tt.profile), func(t *testing.T) {
			for i := 0; i < 20; i++ {
				got := pacingDelay(tt.profile)
				if tt.profile == PacingNone {
					if got != 0 {
						t.Errorf("pacingDelay(%q) = %v, want 0", tt.profile, got)
					}
					return
				}
				if got < tt.minDur || got > tt.maxDur {
					t.Errorf("pacingDelay(%q) = %v, want between %v and %v", tt.profile, got, tt.minDur, tt.maxDur)
				}
			}
		})
	}
}

func TestPacingDelayUnknownProfile(t *testing.T) {
	got := pacingDelay("nonexistent")
	if got != 0 {
		t.Errorf("pacingDelay(unknown) = %v, want 0", got)
	}
}
