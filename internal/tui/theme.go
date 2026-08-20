package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/lib"
)

// Interactive reports whether it is safe to wait for terminal input.
func Interactive() bool {
	if !lib.IsTTY {
		return false
	}
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// Section renders a compact wizard group heading.
func Section(label string) string { return lib.Purple(label) }

// Cursor renders the active-row marker.
func Cursor(active bool) string {
	if active {
		return lib.Green(">")
	}
	return " "
}

// Checkbox renders a state marker that remains understandable without color.
func Checkbox(selected, active bool) string {
	marker := "[ ]"
	if selected {
		marker = "[•]"
	}
	if selected || active {
		return lib.Green(marker)
	}
	return lib.Gray(marker)
}

// Footer renders concise contextual keyboard guidance.
func Footer(text string) string { return lib.Gray(text) }

// Move clamps a cursor after a keyboard movement.
func Move(cursor, total, delta int) int {
	if total <= 0 {
		return 0
	}
	cursor += delta
	if cursor < 0 {
		return 0
	}
	if cursor >= total {
		return total - 1
	}
	return cursor
}

// HandleCommonKey handles navigation shared by selector models. Enter and
// space remain model-specific because their meaning differs by selector.
func HandleCommonKey(msg tea.KeyMsg, cursor *int, total int, help func()) (quit bool, consumed bool) {
	switch msg.Type {
	case tea.KeyUp:
		*cursor = Move(*cursor, total, -1)
		return false, true
	case tea.KeyDown:
		*cursor = Move(*cursor, total, 1)
		return false, true
	case tea.KeyHome:
		*cursor = 0
		return false, true
	case tea.KeyEnd:
		*cursor = Move(total-1, total, 0)
		return false, true
	case tea.KeyCtrlC, tea.KeyEsc:
		return true, true
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == '?' && help != nil {
			help()
			return false, true
		}
	}
	return false, false
}
