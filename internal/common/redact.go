package common

import "github.com/google/uuid"

// redact a secret for log inclusion.
func Redact(secret string) string {

	const (
		minLen = 6
		maxLen = 60
	)

	if secret == "" {
		return ""
	}

	if _, ok := parseUuid(secret); ok {
		return secret[:9] + "{REDACTED}" + secret[23:]
	}

	r := []rune(secret)
	l := len(r)

	if l < minLen {
		return "{REDACTED}"
	}

	if l > maxLen {
		// Big secrets, e.g. certs
		return string(r[:20]) + "{REDACTED}" + string(r[(l-20):])
	}

	l /= 3
	return string(r[:l]) + "{REDACTED}" + string(r[(l*2):])
}

// parseUuid parses a UUID from a string.
// Returns a uuid.UUID and true if successful; else zero UUID and false.
func parseUuid(u string) (uuid.UUID, bool) {

	if id, err := uuid.Parse(u); err != nil {
		return uuid.UUID{}, false
	} else {
		return id, true
	}
}
