package security

import (
	"errors"
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		contains    string
		notContains string
	}{
		{
			name:        "API key in JSON",
			input:       `{"apiKey": "sk-1234567890abcdefghij"}`,
			contains:    "[REDACTED]",
			notContains: "sk-1234567890",
		},
		{
			name:        "Bearer token",
			input:       "Authorization: Bearer abc123xyz789",
			contains:    "Bearer [REDACTED]",
			notContains: "abc123xyz789",
		},
		{
			name:        "Password field",
			input:       "password=secretpass123",
			contains:    "password=[REDACTED]",
			notContains: "secretpass123",
		},
		{
			name:        "OpenAI env style",
			input:       "OPENAI_API_KEY=sk-proj-abcdefghijklmnopqrstuvwxyz",
			contains:    "OPENAI_API_KEY=[REDACTED]",
			notContains: "sk-proj-abcdef",
		},
		{
			name:     "No secrets",
			input:    "This is a normal log message",
			contains: "normal log message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSecrets(tt.input)
			if tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Fatalf("expected output to contain %q, got %q", tt.contains, result)
			}
			if tt.notContains != "" && strings.Contains(result, tt.notContains) {
				t.Fatalf("expected output to NOT contain %q, got %q", tt.notContains, result)
			}
		})
	}
}

func TestRedactAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Long key",
			input:    "sk-1234567890abcdefghij",
			expected: "sk-1...ghij",
		},
		{
			name:     "Short key",
			input:    "abc",
			expected: "[REDACTED]",
		},
		{
			name:     "Empty key",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactAPIKey(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRedactPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "US phone",
			input:    "+1-555-123-4567",
			expected: "***-***-4567",
		},
		{
			name:     "Short number",
			input:    "123",
			expected: "[REDACTED]",
		},
		{
			name:     "Empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactPhone(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	err := errors.New("API error: apiKey=sk-1234567890abcdef")
	result := SanitizeError(err)

	if strings.Contains(result, "sk-1234567890") {
		t.Fatalf("expected secret to be redacted, got %q", result)
	}
	if !strings.Contains(result, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in output, got %q", result)
	}
}
