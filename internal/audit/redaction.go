package audit

import "regexp"

var sensitiveAssignments = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(passwd\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(secret\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(token\s*[=:]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)(credential\s*[=:]\s*)([^\s,;]+)`),
}

// RedactSensitiveText is an audit/logging hook. It must not be applied to raw
// evidence before persistence because Truthwatcher is evidence-first.
func RedactSensitiveText(value string) string {
	for _, pattern := range sensitiveAssignments {
		value = pattern.ReplaceAllString(value, `${1}[REDACTED]`)
	}
	return value
}
