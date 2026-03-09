package eos

import "testing"

func TestRender(t *testing.T) {
	out, err := Driver{}.Render("leaf-1", map[string]string{"hostname": "leaf-1"})
	if err != nil || out == "" { t.Fatalf("expected output") }
}
