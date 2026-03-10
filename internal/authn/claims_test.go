package authn

import (
	"encoding/base64"
	"testing"
)

func TestParserParse(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-1","roles":["admin"]}`))
	token := "x." + payload + ".y"
	claims, err := NewParser().Parse(token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1 got %q", claims.Subject)
	}
}
