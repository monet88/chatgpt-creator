package email

import "testing"

func TestParseSearchResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single message",
			input:    "* SEARCH 42\nA001 OK SEARCH completed",
			expected: []string{"42"},
		},
		{
			name:     "multiple messages",
			input:    "* SEARCH 1 5 10 23\nA001 OK SEARCH completed",
			expected: []string{"1", "5", "10", "23"},
		},
		{
			name:     "no matches",
			input:    "* SEARCH\nA001 OK SEARCH completed",
			expected: nil,
		},
		{
			name:     "empty response",
			input:    "A001 OK SEARCH completed",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseSearchResponse(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("len = %d, want %d", len(result), len(tc.expected))
			}
			for i, id := range result {
				if id != tc.expected[i] {
					t.Errorf("id[%d] = %q, want %q", i, id, tc.expected[i])
				}
			}
		})
	}
}

func TestOTPRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard OTP",
			input:    "Your verification code is 486460. Do not share this code.",
			expected: "486460",
		},
		{
			name:     "OTP at start",
			input:    "123456 is your verification code",
			expected: "123456",
		},
		{
			name:     "no OTP",
			input:    "Hello, your account is ready",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := otpRegex.FindStringSubmatch(tc.input)
			if tc.expected == "" {
				if len(matches) >= 1 {
					t.Fatalf("expected no match, got %q", matches[0])
				}
				return
			}
			if len(matches) < 1 {
				t.Fatal("expected OTP match, got none")
			}
			if matches[0] != tc.expected {
				t.Fatalf("OTP = %q, want %q", matches[0], tc.expected)
			}
		})
	}
}

func TestGeneratorEmailProvider_Close(t *testing.T) {
	p := &GeneratorEmailProvider{}
	if err := p.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
