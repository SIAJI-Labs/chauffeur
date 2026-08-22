package commands

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

type linkSetup struct {
	php               string
	projectType       projects.ProjectType
	proxyPort         int
	secure, dedicated bool
	cancelled         bool
}

type linkWizardModel struct {
	groups          [][]string
	labels          []string
	page            int
	cursor          int
	selected        map[string]bool
	result          linkSetup
	aborted         bool
	proxyInput      bool
	proxyInputValue string
	proxyInputError string
}

func runLinkWizard(path string, cfg workspace.Config, projectType projects.ProjectType) (linkSetup, error) {
	if projectType == projects.TypeReverseProxy {
		defaultPort := projects.DefaultProxyPortFor(path)
		proxyOptions := []string{strconv.Itoa(defaultPort)}
		for _, port := range []int{3000, 5173, 4200, 8080} {
			if port != defaultPort {
				proxyOptions = append(proxyOptions, strconv.Itoa(port))
			}
		}
		proxyOptions = append(proxyOptions, "Custom port")
		model := linkWizardModel{
			groups: [][]string{proxyOptions, {"HTTP", "HTTPS"}},
			labels: []string{"Reverse proxy port", "SSL"},
			result: linkSetup{projectType: projectType},
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
		result: linkSetup{projectType: projectType},
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
	if m.proxyInput {
		switch key.Type {
		case tea.KeyBackspace:
			if len(m.proxyInputValue) > 0 {
				m.proxyInputValue = m.proxyInputValue[:len(m.proxyInputValue)-1]
			}
			m.proxyInputError = ""
		case tea.KeyRunes:
			m.proxyInputValue += string(key.Runes)
			m.proxyInputError = ""
		case tea.KeyEnter:
			port, err := strconv.Atoi(strings.TrimSpace(m.proxyInputValue))
			if err != nil || port < 1 || port > 65535 {
				m.proxyInputError = "Enter a port between 1 and 65535"
				return m, nil
			}
			m.result.proxyPort = port
			m.proxyInput = false
			m.proxyInputValue = ""
			m.proxyInputError = ""
			if m.page < len(m.groups)-1 {
				m.page++
				m.cursor = 0
			}
		case tea.KeyEsc:
			m.proxyInput = false
			m.proxyInputValue = ""
			m.proxyInputError = ""
		}
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
		if m.result.projectType == projects.TypeReverseProxy &&
			m.page == 0 && m.groups[m.page][m.cursor] == "Custom port" {
			m.proxyInput = true
			m.proxyInputValue = ""
			m.proxyInputError = ""
			return m, nil
		}
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
	if m.result.projectType == projects.TypeReverseProxy {
		if m.page == 0 {
			m.result.proxyPort, _ = strconv.Atoi(item)
		} else if m.page == 1 {
			m.result.secure = item == "HTTPS"
		}
		return
	}
	switch m.page {
	case 0:
		m.result.php = strings.Fields(item)[0]
	case 1:
		m.result.secure = item == "HTTPS"
	case 2:
		m.result.dedicated = item == "dedicated FPM"
	}
}

type projectTypeWizardModel struct {
	detected     projects.ProjectType
	options      []string
	cursor       int
	selected     projects.ProjectType
	choosingType bool
	aborted      bool
}

func runProjectTypeWizard(detected projects.ProjectType) (projects.ProjectType, error) {
	model := projectTypeWizardModel{detected: detected}
	if detected == projects.TypeUnknown {
		model.beginTypeSelection()
	} else {
		model.options = []string{
			"Continue with detected " + projectTypeSetupLabel(detected) + " setup",
			"Change project type",
		}
	}
	program := tea.NewProgram(model)
	final, err := program.Run()
	if err != nil {
		return projects.TypeUnknown, err
	}
	result, ok := final.(projectTypeWizardModel)
	if !ok || result.aborted {
		return projects.TypeUnknown, nil
	}
	return result.selected, nil
}

func (m projectTypeWizardModel) Init() tea.Cmd { return nil }

func (m projectTypeWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.Type {
	case tea.KeyUp:
		m.cursor = tui.Move(m.cursor, len(m.options), -1)
	case tea.KeyDown:
		m.cursor = tui.Move(m.cursor, len(m.options), 1)
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = len(m.options) - 1
	case tea.KeyEnter:
		if !m.choosingType {
			if m.cursor == 0 {
				m.selected = m.detected
				return m, tea.Quit
			}
			m.beginTypeSelection()
			return m, nil
		}
		m.selected = projectTypeOptions()[m.cursor]
		return m, tea.Quit
	case tea.KeyCtrlC, tea.KeyEsc:
		m.aborted = true
		return m, tea.Quit
	}
	return m, nil
}

func (m projectTypeWizardModel) View() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s\n\n", lib.Green("❯ project detection"))
	if m.detected == projects.TypeUnknown {
		fmt.Fprintln(&b, "  Project type could not be detected automatically.")
	} else {
		fmt.Fprintf(&b, "  Project detected as %s.\n", projectTypeDetectionLabel(m.detected))
		fmt.Fprintf(&b, "  Using %s setup unless you change it below.\n", projectTypeSetupLabel(m.detected))
	}
	fmt.Fprintln(&b)
	if m.choosingType {
		fmt.Fprintf(&b, "  %s\n\n", lib.Green("❯ choose project type"))
	}
	for i, option := range m.options {
		if i == m.cursor {
			fmt.Fprintf(&b, "  %s %s\n", tui.Cursor(true), lib.Green(option))
		} else {
			fmt.Fprintf(&b, "  %s %s\n", tui.Cursor(false), lib.Gray(option))
		}
	}
	fmt.Fprintf(&b, "\n  %s\n", tui.Footer("enter select · esc cancel"))
	return b.String()
}

func (m *projectTypeWizardModel) beginTypeSelection() {
	m.choosingType = true
	m.options = []string{"Laravel", "WordPress", "PHP", "Reverse proxy"}
	m.cursor = 0
}

func projectTypeOptions() []projects.ProjectType {
	return []projects.ProjectType{
		projects.TypeLaravel,
		projects.TypeWordPress,
		projects.TypePHP,
		projects.TypeReverseProxy,
	}
}

func projectTypeDetectionLabel(projectType projects.ProjectType) string {
	if projectType == projects.TypeReverseProxy {
		return "a JavaScript application"
	}
	return projectTypeSetupLabel(projectType)
}

func projectTypeSetupLabel(projectType projects.ProjectType) string {
	switch projectType {
	case projects.TypeLaravel:
		return "Laravel"
	case projects.TypeWordPress:
		return "WordPress"
	case projects.TypePHP:
		return "PHP"
	case projects.TypeReverseProxy:
		return "reverse proxy"
	default:
		return "manual"
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
	if m.proxyInput {
		fmt.Fprintf(&b, "  Custom port: %s_\n", m.proxyInputValue)
		if m.proxyInputError != "" {
			fmt.Fprintf(&b, "  %s\n", lib.Red(m.proxyInputError))
		}
		fmt.Fprintf(&b, "\n  %s\n", tui.Footer("type a port · enter confirm · esc back"))
		return b.String()
	}
	fmt.Fprintf(&b, "  %s\n", tui.Footer("enter next · space toggle · esc cancel · ? help"))
	return b.String()
}
