package commands

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

type linkSetup struct {
	php               string
	secure, dedicated bool
	cancelled         bool
}

type linkWizardModel struct {
	groups   [][]string
	labels   []string
	page     int
	cursor   int
	selected map[string]bool
	result   linkSetup
	aborted  bool
}

func runLinkWizard(path string, cfg workspace.Config) (linkSetup, error) {
	phpVersions := installers.ListInstalledPHP(workspace.Root())
	if len(phpVersions) == 0 && cfg.PHP.DefaultVersion != "" {
		phpVersions = []string{cfg.PHP.DefaultVersion + " (not installed)"}
	}
	if len(phpVersions) == 0 {
		phpVersions = []string{"Install a PHP version first"}
	}

	model := linkWizardModel{
		groups: [][]string{
			phpVersions,
			{"HTTP", "HTTPS"},
			{"shared FPM", "dedicated FPM"},
		},
		labels: []string{"PHP version", "SSL", "FPM mode"},
	}
	if cfg.PHP.DefaultVersion != "" {
		for i, version := range model.groups[0] {
			if strings.HasPrefix(version, cfg.PHP.DefaultVersion) {
				model.cursor = i
				break
			}
		}
	}
	program := tea.NewProgram(model)
	final, err := program.Run()
	if err != nil {
		return linkSetup{}, err
	}
	result, ok := final.(linkWizardModel)
	if !ok || result.aborted {
		return linkSetup{cancelled: true}, nil
	}
	return result.result, nil
}

func (m linkWizardModel) Init() tea.Cmd { return nil }

func (m linkWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyUp:
		m.cursor = tui.Move(m.cursor, len(m.groups[m.page]), -1)
	case tea.KeyDown:
		m.cursor = tui.Move(m.cursor, len(m.groups[m.page]), 1)
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = tui.Move(len(m.groups[m.page])-1, len(m.groups[m.page]), 0)
	case tea.KeySpace:
	case tea.KeyEnter:
		if m.page < len(m.groups)-1 {
			m.capturePage()
			m.page++
			m.cursor = 0
			return m, nil
		}
		m.capturePage()
		m.result.cancelled = false
		return m, tea.Quit
	case tea.KeyCtrlC, tea.KeyEsc:
		if m.page > 0 {
			m.page--
			m.cursor = 0
			return m, nil
		}
		m.aborted = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *linkWizardModel) capturePage() {
	item := m.groups[m.page][m.cursor]
	switch m.page {
	case 0:
		m.result.php = strings.Fields(item)[0]
	case 1:
		m.result.secure = item == "HTTPS"
	case 2:
		m.result.dedicated = item == "dedicated FPM"
	}
}

func (m linkWizardModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", lib.Green("❯ chauf link"))
	for i, label := range m.labels {
		active := i == m.page
		heading := lib.Gray(label)
		if active {
			heading = tui.Section(label)
		}
		fmt.Fprintf(&b, "  %s\n", heading)
		for j, item := range m.groups[i] {
			focused := active && j == m.cursor
			cursor := tui.Cursor(focused)
			if focused {
				fmt.Fprintf(&b, "  %s %s\n", cursor, lib.Green(item))
			} else {
				fmt.Fprintf(&b, "  %s %s\n", cursor, lib.Gray(item))
			}
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "  %s\n", tui.Footer("enter next · space toggle · esc cancel · ? help"))
	return b.String()
}
