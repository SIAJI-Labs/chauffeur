package tui

import "testing"

func TestMoveClampsCursor(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		total  int
		delta  int
		want   int
	}{
		{"empty", 4, 0, 1, 0},
		{"top", 0, 3, -1, 0},
		{"bottom", 2, 3, 1, 2},
		{"middle", 1, 4, 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Move(tt.cursor, tt.total, tt.delta); got != tt.want {
				t.Fatalf("Move(%d, %d, %d) = %d, want %d", tt.cursor, tt.total, tt.delta, got, tt.want)
			}
		})
	}
}

func TestMarkersPreserveStateWithoutColor(t *testing.T) {
	if got := Cursor(true); got != ">" {
		t.Fatalf("Cursor(true) = %q, want >", got)
	}
	if got := Checkbox(true, false); got != "[•]" {
		t.Fatalf("Checkbox(selected) = %q, want [•]", got)
	}
	if got := Checkbox(false, false); got != "[ ]" {
		t.Fatalf("Checkbox(unselected) = %q, want [ ]", got)
	}
}
