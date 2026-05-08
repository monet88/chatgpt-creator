package register

import (
	"net/url"
	"regexp"
	"strings"
)

var tokenLikePattern = regexp.MustCompile(`(?i)(token|cookie|authorization|password)\s*[:=]\s*[^\s,;]+`)

// RedactPassword returns "[redacted]" for non-empty passwords or "(random)" for empty.
func RedactPassword(password string) string {
	if password == "" {
		return "(random)"
	}
	return "[redacted]"
}

// RedactProxy strips userinfo credentials from proxy URLs.
func RedactProxy(proxy string) string {
	if proxy == "" {
		return ""
	}
	parsed, err := url.Parse(proxy)
	if err != nil {
		return "[redacted]"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("[redacted]", "[redacted]")
	}
	return parsed.String()
}

func sanitizeLogLine(input string) string {
	cleaned := strings.ReplaceAll(input, "\n", "\\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\\r")
	return strings.TrimSpace(cleaned)
}

func redactSensitiveText(input string) string {
	return tokenLikePattern.ReplaceAllStringFunc(input, func(_ string) string {
		return "[redacted]"
	})
}

func safeLogMessage(input string) string {
	return redactSensitiveText(sanitizeLogLine(input))
}
