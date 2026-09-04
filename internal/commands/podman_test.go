package commands

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMultiSelectModelView_ShowsCursorWithinViewport(t *testing.T) {
	choices := make([]string, 12)
	for i := range choices {
		choices[i] = fmt.Sprintf("db-%02d", i)
	}

	model := multiSelectModel{
		choices:  choices,
		selected: make(map[int]bool),
		title:    "Select databases to backup",
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m, ok := updated.(multiSelectModel)
	if !ok {
		t.Fatalf("expected multiSelectModel after window size update, got %T", updated)
	}

	for i := 0; i < 8; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m, ok = updated.(multiSelectModel)
		if !ok {
			t.Fatalf("expected multiSelectModel after key update, got %T", updated)
		}
	}

	view := m.View()

	if !strings.Contains(view, "> [ ] db-08") {
		t.Fatalf("expected viewport to include cursor item db-08, got view:\n%s", view)
	}

	if strings.Contains(view, "db-00") {
		t.Fatalf("expected viewport to scroll past db-00 for short terminal, got view:\n%s", view)
	}
	if strings.Contains(view, "db-11") {
		t.Fatalf("expected viewport height to limit visible choices, got view:\n%s", view)
	}

	lineCount := len(strings.Split(view, "\n"))
	if lineCount > 10 {
		t.Fatalf("expected rendered view to respect window height, got %d lines in view:\n%s", lineCount, view)
	}
}

func TestSingleSelectModelView_ShowsCursorWithinViewport(t *testing.T) {
	items := make([]string, 12)
	for i := range items {
		items[i] = fmt.Sprintf("service-%02d", i)
	}

	model := singleSelectModel{
		items: items,
		title: "Select a service",
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m, ok := updated.(singleSelectModel)
	if !ok {
		t.Fatalf("expected singleSelectModel after window size update, got %T", updated)
	}

	for i := 0; i < 8; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m, ok = updated.(singleSelectModel)
		if !ok {
			t.Fatalf("expected singleSelectModel after key update, got %T", updated)
		}
	}

	view := m.View()

	if !strings.Contains(view, "> service-08") {
		t.Fatalf("expected viewport to include cursor item service-08, got view:\n%s", view)
	}

	if strings.Contains(view, "service-00") {
		t.Fatalf("expected viewport to scroll past service-00 for short terminal, got view:\n%s", view)
	}
	if strings.Contains(view, "service-11") {
		t.Fatalf("expected viewport height to limit visible items, got view:\n%s", view)
	}

	lineCount := len(strings.Split(view, "\n"))
	if lineCount > 10 {
		t.Fatalf("expected rendered view to respect window height, got %d lines in view:\n%s", lineCount, view)
	}
}

func TestContainerSelectModelView_ShowsCursorWithinViewport(t *testing.T) {
	containers := make([]containerInfo, 12)
	for i := range containers {
		containers[i] = containerInfo{
			Name:   fmt.Sprintf("container-%02d", i),
			Engine: "docker",
			Status: "running",
		}
	}

	model := containerSelectModel{
		containers: containers,
		selected:   make(map[int]bool),
		title:      "Select containers",
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m, ok := updated.(containerSelectModel)
	if !ok {
		t.Fatalf("expected containerSelectModel after window size update, got %T", updated)
	}

	for i := 0; i < 8; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m, ok = updated.(containerSelectModel)
		if !ok {
			t.Fatalf("expected containerSelectModel after key update, got %T", updated)
		}
	}

	view := m.View()

	if !strings.Contains(view, "> [ ] container-08") {
		t.Fatalf("expected viewport to include cursor item container-08, got view:\n%s", view)
	}

	if strings.Contains(view, "container-00") {
		t.Fatalf("expected viewport to scroll past container-00 for short terminal, got view:\n%s", view)
	}
	if strings.Contains(view, "container-11") {
		t.Fatalf("expected viewport height to limit visible containers, got view:\n%s", view)
	}

	lineCount := len(strings.Split(view, "\n"))
	if lineCount > 10 {
		t.Fatalf("expected rendered view to respect window height, got %d lines in view:\n%s", lineCount, view)
	}
}

func TestGenericPodmanCommands(t *testing.T) {
	for _, command := range []string{"ps", "list", "inspect", "logs", "exec"} {
		if !isGenericPodmanCommand(command) {
			t.Errorf("isGenericPodmanCommand(%q) = false", command)
		}
	}
	if isGenericPodmanCommand("create") {
		t.Error("database command must not be treated as generic Podman")
	}
}

func TestGenericPodmanOnlyAcceptsChauffeurContainers(t *testing.T) {
	if !isChauffeurContainer("chauf-postgres") {
		t.Fatal("expected Chauffeur-prefixed container to be accepted")
	}
	if isChauffeurContainer("unrelated-container") {
		t.Fatal("expected unrelated container to be rejected")
	}
}
