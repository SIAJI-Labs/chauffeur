package tui

import "testing"

func TestConfirmTextRequiresInteractiveTerminal(t *testing.T) {
	if ConfirmText("Type yes", "yes") {
		t.Fatal("ConfirmText should refuse when stdin is not interactive")
	}
}
