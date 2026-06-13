package audit

import (
	"strings"
	"testing"
)

func TestRedactSensitiveText(t *testing.T) {
	input := "username=netops password=supersecret token:abc123 secret = hidden credential=local-ref"
	got := RedactSensitiveText(input)

	for _, leaked := range []string{"supersecret", "abc123", "hidden", "local-ref"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted output leaked %q: %s", leaked, got)
		}
	}
	if strings.Count(got, "[REDACTED]") != 4 {
		t.Fatalf("redacted output = %q, want 4 redactions", got)
	}
}
