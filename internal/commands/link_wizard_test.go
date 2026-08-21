package commands

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLinkWizardCapturesSelections(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"8.3"}, {"HTTP", "HTTPS"}, {"shared FPM", "dedicated FPM"}},
		labels: []string{"PHP version", "SSL", "FPM mode"},
	}
	for page, cursor := range []int{0, 1, 1} {
		m.page, m.cursor = page, cursor
		m.capturePage()
	}
	if m.result.php != "8.3" {
		t.Fatalf("unexpected setup selection: %+v", m.result)
	}
	if !m.result.secure || !m.result.dedicated {
		t.Fatalf("unexpected runtime selection: %+v", m.result)
	}
}

func TestLinkWizardViewShowsSharedVisualLanguage(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"8.3"}, {"HTTP", "HTTPS"}, {"shared FPM", "dedicated FPM"}},
		labels: []string{"PHP version", "SSL", "FPM mode"},
		page:   1,
	}
	view := m.View()
	for _, want := range []string{"PHP version", ">", "SSL", "enter next"} {
		if !strings.Contains(view, want) {
			t.Fatalf("wizard view missing %q:\n%s", want, view)
		}
	}
}

func TestLinkWizardEscCancelsWithoutApplying(t *testing.T) {
	m := linkWizardModel{groups: [][]string{{"8.3"}}, labels: []string{"PHP version"}}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := updated.(linkWizardModel)
	if !got.aborted || cmd == nil {
		t.Fatalf("escape = aborted %t, cmd nil %t; want abort and quit", got.aborted, cmd == nil)
	}
}
