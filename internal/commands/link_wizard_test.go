package commands

import (
	"os"
	"path/filepath"
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

func TestLinkWizardReviewShowsCompletePreviewWithoutApplying(t *testing.T) {
	m := linkWizardModel{
		groups:        [][]string{{"8.3"}},
		labels:        []string{"PHP version"},
		page:          0,
		review:        true,
		path:          "/work/shop",
		slug:          "shop",
		documentRoot:  "/work/shop/public",
		runtimeEngine: "podman",
		result:        linkSetup{php: "8.3", domain: "shop.test", projectType: projects.TypeLaravel, secure: true},
	}
	view := m.View()
	for _, want := range []string{"/work/shop", "shop", "/work/shop/public", "Runtime:   podman", "changes to apply", "prepare podman PHP-FPM runtime", "warnings:", "resources untouched", "waiting for explicit confirmation", "no changes have been applied"} {
		if !strings.Contains(view, want) {
			t.Fatalf("review missing %q:\n%s", want, view)
		}
	}
	if m.result.confirmed {
		t.Fatal("rendering review must not confirm apply")
	}
}

func TestLinkWizardCtrlCCancelsFromReview(t *testing.T) {
	m := linkWizardModel{review: true, result: linkSetup{php: "8.3"}}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	got := updated.(linkWizardModel)
	if !got.aborted || cmd == nil || got.result.confirmed {
		t.Fatalf("ctrl+c review result = %+v, cmd nil %t; want cancellation", got, cmd == nil)
	}
}

func TestLinkWizardCancellationLeavesFilesystemUntouched(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "composer.json")
	if err := os.WriteFile(marker, []byte(`{"require":{"php":"^8.3"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	m := linkWizardModel{
		groups: [][]string{{"8.3"}},
		labels: []string{"PHP version"},
		page:   0,
		path:   root,
		result: linkSetup{php: "8.3", projectType: projects.TypePHP},
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || !m.review {
		t.Fatal("enter should show review before mutation")
	}
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(linkWizardModel)
	if cmd == nil || !m.aborted || m.result.confirmed {
		t.Fatalf("cancel result = %+v; want aborted and unconfirmed", m.result)
	}
	after, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("wizard cancellation changed project files")
	}
}

func TestLinkWizardReviewEnterConfirmsExactlyOnce(t *testing.T) {
	m := linkWizardModel{review: true, result: linkSetup{php: "8.3"}}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(linkWizardModel)
	if !got.result.confirmed || cmd == nil {
		t.Fatalf("enter review result = %+v, cmd nil %t; want confirmation", got.result, cmd == nil)
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

func TestLinkWizardAcceptsCustomDomain(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"shop.test", "Custom domain"}, {"HTTP"}},
		labels: []string{"Primary domain", "SSL"},
		result: linkSetup{domain: "shop.test"},
		cursor: 1,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || !m.domainInput {
		t.Fatalf("domain input = %t, cmd nil %t", m.domainInput, cmd == nil)
	}
	m.domainValue = "admin.test"
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || m.domainInput || m.result.domain != "admin.test" || m.page != 1 {
		t.Fatalf("domain result = %+v, page %d, input %t", m.result, m.page, m.domainInput)
	}
}

func TestLinkWizardAcceptsAndPreviewsAliases(t *testing.T) {
	m := linkWizardModel{
		groups: [][]string{{"shop.test", "Custom domain"}, {"No aliases", "Custom aliases"}, {"HTTP"}},
		labels: []string{"Primary domain", "Aliases", "SSL"},
		result: linkSetup{domain: "shop.test"},
		page:   1,
		cursor: 1,
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || !m.aliasInput {
		t.Fatalf("alias input = %t, cmd nil %t; want input", m.aliasInput, cmd == nil)
	}
	m.aliasValue = "admin.shop.test, api.shop.test"
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(linkWizardModel)
	if cmd != nil || m.aliasInput || len(m.result.aliases) != 2 || m.page != 2 {
		t.Fatalf("alias result = %+v, page %d, input %t", m.result, m.page, m.aliasInput)
	}
	m.review = true
	view := m.View()
	if !strings.Contains(view, "admin.shop.test, api.shop.test") {
		t.Fatalf("review missing aliases:\n%s", view)
	}
}

func TestLinkWizardRejectsDuplicateAliases(t *testing.T) {
	m := linkWizardModel{aliasInput: true, aliasValue: "admin.test, admin.test"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(linkWizardModel)
	if cmd != nil || got.aliasError == "" || len(got.result.aliases) != 0 {
		t.Fatalf("duplicate aliases result = %+v, cmd nil %t; want validation error", got, cmd == nil)
	}
}
