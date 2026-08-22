package commands

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/projects"
)

func TestParseProjectType(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  projects.ProjectType
	}{
		{"laravel", projects.TypeLaravel},
		{"wordpress", projects.TypeWordPress},
		{"php", projects.TypePHP},
		{"reverse-proxy", projects.TypeReverseProxy},
	} {
		got, err := parseProjectType(tc.input)
		if err != nil || got != tc.want {
			t.Errorf("parseProjectType(%q) = %q, %v; want %q", tc.input, got, err, tc.want)
		}
	}
}

func TestProjectTypeWizardSelectsReverseProxy(t *testing.T) {
	m := projectTypeWizardModel{
		detected:     projects.TypeUnknown,
		options:      []string{"Laravel", "WordPress", "PHP", "Reverse proxy"},
		choosingType: true,
		cursor:       3,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(projectTypeWizardModel)
	if got.selected != projects.TypeReverseProxy || cmd == nil {
		t.Fatalf("selection = %q, cmd nil %t", got.selected, cmd == nil)
	}
}

func TestProjectTypeWizardShowsDetectionAndAllowsChange(t *testing.T) {
	m := projectTypeWizardModel{detected: projects.TypeReverseProxy, cursor: 1}
	m.options = []string{
		"Continue with detected reverse proxy setup",
		"Change project type",
	}
	view := m.View()
	for _, want := range []string{"Project detected as a JavaScript application", "Using reverse proxy setup"} {
		if !strings.Contains(view, want) {
			t.Fatalf("detection view missing %q:\n%s", want, view)
		}
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	changed := updated.(projectTypeWizardModel)
	if cmd != nil || !changed.choosingType || len(changed.options) != 4 {
		t.Fatalf("change selection = choosing %t, options %d, cmd nil %t", changed.choosingType, len(changed.options), cmd == nil)
	}
}

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

func TestReverseProxyWizardCapturesPortAndSSL(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"3000", "5173"}, {"HTTP", "HTTPS"}},
		labels: []string{"Reverse proxy port", "SSL"},
		result: linkSetup{projectType: projects.TypeReverseProxy},
	}
	m.page, m.cursor = 0, 1
	m.capturePage()
	m.page, m.cursor = 1, 1
	m.capturePage()
	if m.result.proxyPort != 5173 || !m.result.secure {
		t.Fatalf("unexpected reverse proxy setup: %+v", m.result)
	}
}

func TestReverseProxyWizardAcceptsCustomPort(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"3000", "Custom port"}, {"HTTP", "HTTPS"}},
		labels: []string{"Reverse proxy port", "SSL"},
		result: linkSetup{projectType: projects.TypeReverseProxy},
		cursor: 1,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || !m.proxyInput {
		t.Fatalf("custom port entry = input %t, cmd nil %t", m.proxyInput, cmd == nil)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9123")})
	m = updated.(linkWizardModel)
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || m.result.proxyPort != 9123 || m.page != 1 {
		t.Fatalf("custom port result = %+v, page %d, cmd nil %t", m.result, m.page, cmd == nil)
	}
}
