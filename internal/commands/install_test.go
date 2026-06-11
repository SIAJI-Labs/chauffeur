package commands

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPHPSelectModelView_ShowsCursorWithinViewportWithInstalledRows(t *testing.T) {
	items := make([]string, 12)
	installed := map[string]bool{}

	for i := range items {
		items[i] = fmt.Sprintf("8.%d", i)
	}

	installed[items[1]] = true
	installed[items[4]] = true
	installed[items[9]] = true

	model := phpSelectModel{
		items:     items,
		title:     "Select PHP version",
		installed: installed,
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m, ok := updated.(phpSelectModel)
	if !ok {
		t.Fatalf("expected phpSelectModel after window size update, got %T", updated)
	}

	for i := 0; i < 5; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m, ok = updated.(phpSelectModel)
		if !ok {
			t.Fatalf("expected phpSelectModel after key update, got %T", updated)
		}
	}

	view := m.View()

	if !strings.Contains(view, "↑ more") {
		t.Fatalf("expected clipped viewport to show top indicator, got view:\n%s", view)
	}
	if !strings.Contains(view, "↓ more") {
		t.Fatalf("expected clipped viewport to show bottom indicator, got view:\n%s", view)
	}
	if !strings.Contains(view, "(installed)") {
		t.Fatalf("expected installed rows to remain rendered with installed styling, got view:\n%s", view)
	}
	if !strings.Contains(view, "> 8.7") {
		t.Fatalf("expected viewport to include active selectable item 8.7, got view:\n%s", view)
	}

	if strings.Contains(view, "8.0") {
		t.Fatalf("expected viewport to scroll past earliest rows for short terminal, got view:\n%s", view)
	}
	if strings.Contains(view, "8.11") {
		t.Fatalf("expected viewport height to limit visible rows, got view:\n%s", view)
	}

	lineCount := len(strings.Split(view, "\n"))
	if lineCount > 10 {
		t.Fatalf("expected rendered view to respect window height, got %d lines in view:\n%s", lineCount, view)
	}
}

func TestPHPSelectModelView_ShowsViewportIndicatorsWhenClipped(t *testing.T) {
	items := make([]string, 12)
	installed := map[string]bool{}

	for i := range items {
		items[i] = fmt.Sprintf("8.%d", i)
	}

	installed[items[1]] = true
	installed[items[4]] = true
	installed[items[9]] = true

	model := phpSelectModel{
		items:     items,
		title:     "Select PHP version",
		installed: installed,
		height:    10,
		cursor:    5,
	}

	view := model.View()

	if !strings.Contains(view, "↑ more") {
		t.Fatalf("expected clipped viewport to show top indicator, got view:\n%s", view)
	}
	if !strings.Contains(view, "↓ more") {
		t.Fatalf("expected clipped viewport to show bottom indicator, got view:\n%s", view)
	}
	if !strings.Contains(view, "> 8.7") {
		t.Fatalf("expected active selectable item to remain visible, got view:\n%s", view)
	}
	if !strings.Contains(view, "(installed)") {
		t.Fatalf("expected installed row styling to remain visible in clipped viewport, got view:\n%s", view)
	}
	if lineCount := len(strings.Split(view, "\n")); lineCount > 10 {
		t.Fatalf("expected rendered view to respect window height, got %d lines in view:\n%s", lineCount, view)
	}
}
