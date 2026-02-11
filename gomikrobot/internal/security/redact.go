// Package security provides security utilities for GoMikroBot.
package security

import (
	"regexp"
	"strings"
)

var (
	// keyValueSecretRegex matches common key/value secret patterns, capturing:
	//  1) key name
	//  2) separator (":", "=", or whitespace)
	//  3) value
	keyValueSecretRegex = regexp.MustCompile(`(?i)\b(api[_-]?key|token|secret|password|auth)\b(\s*[:=]\s*|\s+)([^\s"']+)`)

	// bearerRegex matches Authorization header values.
	bearerRegex = regexp.MustCompile(`(?i)\bBearer\s+([A-Za-z0-9\-_\.]+)`)

	// Common API key formats (best-effort)
	openAIKeyRegex = regexp.MustCompile(`\bsk-[A-Za-z0-9]{20,}\b`)
	groqKeyRegex   = regexp.MustCompile(`\bgsk_[A-Za-z0-9]{20,}\b`)
)

// RedactSecrets replaces sensitive data in strings with [REDACTED].
// Best-effort: it will not catch every possible secret format.
func RedactSecrets(input string) string {
	out := input
	out = keyValueSecretRegex.ReplaceAllString(out, `$1$2[REDACTED]`)
	out = bearerRegex.ReplaceAllString(out, `Bearer [REDACTED]`)
	out = openAIKeyRegex.ReplaceAllString(out, `[REDACTED]`)
	out = groqKeyRegex.ReplaceAllString(out, `[REDACTED]`)
	return out
}

// RedactAPIKey redacts an API key while leaving a small prefix/suffix for debugging.
func RedactAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "[REDACTED]"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// RedactPhone redacts phone numbers for PII protection (keeps last 4 digits).
func RedactPhone(phone string) string {
	if phone == "" {
		return ""
	}
	digits := regexp.MustCompile(`\d`).FindAllString(phone, -1)
	if len(digits) < 4 {
		return "[REDACTED]"
	}
	return "***-***-" + strings.Join(digits[len(digits)-4:], "")
}

// SanitizeError removes sensitive info from error messages.
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return RedactSecrets(err.Error())
}
