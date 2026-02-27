package ui

import "testing"

// TestKeymapHelp verifies that ShortHelp and FullHelp return bindings.
// This ensures help rendering has consistent data.
func TestKeymapHelp(t *testing.T) {
	km := newKeyMap()
	if len(km.ShortHelp()) == 0 {
		t.Fatalf("expected short help bindings")
	}
	if len(km.FullHelp()) == 0 {
		t.Fatalf("expected full help bindings")
	}
}
