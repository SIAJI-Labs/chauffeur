package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siegg/chauffeur/internal/installers"
	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/projects"
	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
	"github.com/siegg/chauffeur/internal/tui"
	"github.com/siegg/chauffeur/internal/workspace"
)

type linkSetup struct {
	php               string
	domain            string
	projectType       projects.ProjectType
	proxyPort         int
	secure, dedicated bool
	cancelled         bool
	confirmed         bool
	aliases           []string
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
	domainInput     bool
	domainValue     string
	domainError     string
	aliasInput      bool
	aliasValue      string
	aliasError      string
	review          bool
	path            string
	slug            string
	documentRoot    string
	runtimeEngine   string
}

func runLinkWizard(path string, cfg workspace.Config, projectType projects.ProjectType, existing *projects.Project) (linkSetup, error) {
	domain := projects.DomainFromSlug(projects.GenerateSlug(path), cfg.DNS.TLD)
	if existing != nil && existing.Domain != "" {
		domain = existing.Domain
	}
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
			groups: [][]string{proxyOptions, {domain, "Custom domain"}, {"No aliases", "Custom aliases"}, {"HTTP", "HTTPS"}},
			labels: []string{"Reverse proxy port", "Primary domain", "Aliases", "SSL"},
			result: linkSetup{projectType: projectType, domain: domain},
			path:   path, slug: projects.GenerateSlug(path), documentRoot: projects.DocumentRoot(path, projectType), runtimeEngine: cfg.Runtime.Engine,
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

	var phpVersions []string
	if cfg.Runtime.Engine == "podman" {
		for _, version := range []string{"7.4", "8.0", "8.3", "8.5"} {
			result, err := (chauftruntime.ExecRunner{}).Run(context.Background(), "image", "exists", chauftruntime.PHPImage(version))
			if err == nil && result.ExitCode == 0 {
				phpVersions = append(phpVersions, version)
			} else {
				phpVersions = append(phpVersions, version+" (unavailable)")
			}
		}
	} else {
		phpVersions = installers.ListInstalledPHP(workspace.Root())
	}
	if len(phpVersions) == 0 && cfg.PHP.DefaultVersion != "" {
		phpVersions = []string{cfg.PHP.DefaultVersion + " (not installed)"}
	}
	if len(phpVersions) == 0 {
		phpVersions = []string{"Install a PHP version first"}
	}

	model := linkWizardModel{
		groups: [][]string{
			phpVersions,
			{domain, "Custom domain"},
			{"No aliases", "Custom aliases"},
			{"HTTP", "HTTPS"},
			{"shared FPM", "dedicated FPM"},
		},
		labels: []string{"PHP version", "Primary domain", "Aliases", "SSL", "FPM mode"},
		result: linkSetup{projectType: projectType, domain: domain},
		path:   path, slug: projects.GenerateSlug(path), documentRoot: projects.DocumentRoot(path, projectType), runtimeEngine: cfg.Runtime.Engine,
	}
	if existing != nil {
		model.result.php = existing.PHPVersion
		model.result.aliases = append([]string(nil), existing.Aliases...)
		model.result.secure = existing.SSL
		model.result.dedicated = existing.FPM.Dedicated
	}
	if cfg.PHP.DefaultVersion != "" {
		for i, version := range model.groups[0] {
			if strings.HasPrefix(version, cfg.PHP.DefaultVersion) {
				model.cursor = i
				break
			}
		}
	}
	if existing != nil {
		for i, option := range model.groups[1] {
			if option == existing.Domain {
				model.page = 1
				model.cursor = i
				break
			}
		}
		if existing.SSL {
			model.page = 3
			model.cursor = 1
		}
		if existing.FPM.Dedicated {
			model.page = 4
			model.cursor = 1
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
	if key.Type == tea.KeyCtrlC {
		m.aborted = true
		return m, tea.Quit
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
	if m.domainInput {
		switch key.Type {
		case tea.KeyBackspace:
			if len(m.domainValue) > 0 {
				m.domainValue = m.domainValue[:len(m.domainValue)-1]
			}
			m.domainError = ""
		case tea.KeyRunes:
			m.domainValue += string(key.Runes)
			m.domainError = ""
		case tea.KeyEnter:
			domain := strings.TrimSpace(m.domainValue)
			if !projects.IsValidDomain(domain) {
				m.domainError = "Enter a valid .test domain"
				return m, nil
			}
			m.result.domain = domain
			m.domainInput = false
			m.domainValue = ""
			m.domainError = ""
			if m.page < len(m.groups)-1 {
				m.page++
				m.cursor = 0
			}
		case tea.KeyEsc:
			m.domainInput = false
			m.domainValue = ""
			m.domainError = ""
		}
		return m, nil
	}
	if m.aliasInput {
		switch key.Type {
		case tea.KeyBackspace:
			if len(m.aliasValue) > 0 {
				m.aliasValue = m.aliasValue[:len(m.aliasValue)-1]
			}
			m.aliasError = ""
		case tea.KeyRunes:
			m.aliasValue += string(key.Runes)
			m.aliasError = ""
		case tea.KeyEnter:
			var aliases []string
			for _, raw := range strings.Split(m.aliasValue, ",") {
				alias := strings.TrimSpace(raw)
				if alias == "" {
					continue
				}
				if !projects.IsValidDomain(alias) || sliceContains(aliases, alias) {
					m.aliasError = "Enter unique valid .test aliases separated by commas"
					return m, nil
				}
				aliases = append(aliases, alias)
			}
			m.result.aliases = aliases
			m.aliasInput, m.aliasValue, m.aliasError = false, "", ""
			if m.page < len(m.groups)-1 {
				m.page++
				m.cursor = 0
			}
		case tea.KeyEsc:
			m.aliasInput, m.aliasValue, m.aliasError = false, "", ""
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
		if m.review {
			m.result.cancelled = false
			m.result.confirmed = true
			return m, tea.Quit
		}
		if m.result.projectType == projects.TypeReverseProxy &&
			m.page == 0 && m.groups[m.page][m.cursor] == "Custom port" {
			m.proxyInput = true
			m.proxyInputValue = ""
			m.proxyInputError = ""
			return m, nil
		}
		if m.labels[m.page] == "Primary domain" && m.groups[m.page][m.cursor] == "Custom domain" {
			m.domainInput = true
			m.domainValue = m.result.domain
			m.domainError = ""
			return m, nil
		}
		if m.labels[m.page] == "Aliases" && m.groups[m.page][m.cursor] == "Custom aliases" {
			m.aliasInput = true
			m.aliasValue = strings.Join(m.result.aliases, ", ")
			m.aliasError = ""
			return m, nil
		}
		if m.page < len(m.groups)-1 {
			m.capturePage()
			m.page++
			m.cursor = 0
			return m, nil
		}
		m.capturePage()
		m.review = true
		return m, nil
	case tea.KeyEsc:
		if m.domainInput {
			m.domainInput = false
			m.domainValue = ""
			m.domainError = ""
			return m, nil
		}
		if m.review {
			m.review = false
			m.page = len(m.groups) - 1
			m.cursor = 0
			return m, nil
		}
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
		if len(m.labels) == 2 {
			if m.page == 0 {
				m.result.proxyPort, _ = strconv.Atoi(item)
			} else if m.page == 1 {
				m.result.secure = item == "HTTPS"
			}
			return
		}
		switch m.page {
		case 0:
			m.result.proxyPort, _ = strconv.Atoi(item)
		case 1:
			m.result.domain = item
		case 2:
			if item == "No aliases" {
				m.result.aliases = nil
			}
		case 3:
			m.result.secure = item == "HTTPS"
		}
		return
	}
	switch m.labels[m.page] {
	case "PHP version":
		m.result.php = strings.Fields(item)[0]
	case "Primary domain":
		m.result.domain = item
	case "Aliases":
		if item == "No aliases" {
			m.result.aliases = nil
		}
	case "SSL":
		m.result.secure = item == "HTTPS"
	case "FPM mode":
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
	if m.review {
		fmt.Fprintf(&b, "  %s\n\n", tui.Section("review and apply"))
		fmt.Fprintf(&b, "  Path:      %s\n", m.path)
		fmt.Fprintf(&b, "  Slug:      %s\n", m.slug)
		fmt.Fprintf(&b, "  Type:      %s\n", projectTypeSetupLabel(m.result.projectType))
		fmt.Fprintf(&b, "  Root:      %s\n", m.documentRoot)
		fmt.Fprintf(&b, "  PHP:       %s\n", m.result.php)
		if m.result.projectType != projects.TypeReverseProxy {
			fmt.Fprintf(&b, "  Runtime:   %s · PHP-FPM readiness checked at apply\n", m.runtimeEngine)
		}
		fmt.Fprintf(&b, "  Domain:    %s\n", m.result.domain)
		if len(m.result.aliases) > 0 {
			fmt.Fprintf(&b, "  Aliases:   %s\n", strings.Join(m.result.aliases, ", "))
		} else {
			fmt.Fprintln(&b, "  Aliases:   none")
		}
		fmt.Fprintf(&b, "  SSL:       %t\n", m.result.secure)
		if m.result.projectType == projects.TypeReverseProxy {
			fmt.Fprintf(&b, "  Proxy port: %d\n", m.result.proxyPort)
		} else if m.result.dedicated {
			fmt.Fprintln(&b, "  FPM:       dedicated")
		} else {
			fmt.Fprintln(&b, "  FPM:       shared")
		}
		fmt.Fprintln(&b, "\n  changes to apply")
		fmt.Fprintln(&b, "  - write project intent")
		fmt.Fprintln(&b, "  - generate and enable nginx route")
		if m.result.projectType != projects.TypeReverseProxy {
			fmt.Fprintf(&b, "  - prepare %s PHP-FPM runtime\n", m.runtimeEngine)
		}
		if m.result.secure {
			fmt.Fprintln(&b, "  - generate SSL certificate and mount it read-only")
		}
		if m.result.secure && (!projects.MkcertInstalled() || !projects.MkcertCAInstalled()) {
			fmt.Fprintln(&b, "  warnings: mkcert trust prerequisites are unavailable; SSL may fall back to HTTP")
		} else {
			fmt.Fprintln(&b, "  warnings: none detected during read-only review")
		}
		fmt.Fprintln(&b, "  resources untouched: existing projects, containers, certificates, and routes")
		fmt.Fprintln(&b, "  apply state: waiting for explicit confirmation")
		fmt.Fprintln(&b, "\n  no changes have been applied")
		fmt.Fprintf(&b, "\n  %s\n", tui.Footer("enter apply · esc back · ctrl+c cancel"))
		return b.String()
	}
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
			selected := tui.Checkbox(m.optionSelected(i, item), focused)
			displayItem := m.optionDisplay(i, item)
			if focused {
				fmt.Fprintf(&b, "  %s %s %s\n", cursor, selected, lib.Green(displayItem))
			} else {
				fmt.Fprintf(&b, "  %s %s %s\n", cursor, selected, lib.Gray(displayItem))
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
	if m.domainInput {
		fmt.Fprintf(&b, "  Custom domain: %s_\n", m.domainValue)
		if m.domainError != "" {
			fmt.Fprintf(&b, "  %s\n", lib.Red(m.domainError))
		}
		fmt.Fprintf(&b, "\n  %s\n", tui.Footer("type a domain · enter confirm · esc back"))
		return b.String()
	}
	if m.aliasInput {
		fmt.Fprintf(&b, "  Custom aliases: %s_\n", m.aliasValue)
		if m.aliasError != "" {
			fmt.Fprintf(&b, "  %s\n", lib.Red(m.aliasError))
		}
		fmt.Fprintf(&b, "\n  %s\n", tui.Footer("type comma-separated aliases · enter confirm · esc back"))
		return b.String()
	}
	fmt.Fprintf(&b, "  %s\n", tui.Footer("enter next · space toggle · esc cancel · ? help"))
	return b.String()
}

// optionSelected keeps the wizard's committed value visible while the cursor
// moves through later sections. For the current section, the cursor represents
// the pending default until the user confirms it.
func (m linkWizardModel) optionSelected(page int, item string) bool {
	if page >= len(m.labels) || page >= len(m.groups) {
		return false
	}
	if page == m.page {
		return m.cursor == indexOf(m.groups[page], item)
	}
	switch m.labels[page] {
	case "PHP version":
		return m.result.php != "" && strings.HasPrefix(item, m.result.php)
	case "Primary domain":
		return m.result.domain == item
	case "Aliases":
		if item == "No aliases" {
			return len(m.result.aliases) == 0
		}
		return len(m.result.aliases) > 0
	case "SSL":
		return (item == "HTTPS") == m.result.secure
	case "FPM mode":
		return (item == "dedicated FPM") == m.result.dedicated
	case "Reverse proxy port":
		return strconv.Itoa(m.result.proxyPort) == item
	}
	return false
}

func indexOf(items []string, value string) int {
	for i, item := range items {
		if item == value {
			return i
		}
	}
	return -1
}

func (m linkWizardModel) optionDisplay(page int, item string) string {
	if page < len(m.labels) && m.labels[page] == "Aliases" && item == "Custom aliases" && len(m.result.aliases) > 0 {
		return item + " (" + strings.Join(m.result.aliases, ", ") + ")"
	}
	return item
}
