package commands

import (
	"context"
	"testing"

	chauftruntime "github.com/siegg/chauffeur/internal/runtime"
)

type runtimeSelectionRunner struct {
	available map[string]bool
}

func (r runtimeSelectionRunner) Run(_ context.Context, args ...string) (chauftruntime.CommandResult, error) {
	image := args[len(args)-1]
	if r.available[image] {
		return chauftruntime.CommandResult{ExitCode: 0}, nil
	}
	return chauftruntime.CommandResult{ExitCode: 1}, nil
}

func TestInstalledPHPForRuntimeUsesPodmanImages(t *testing.T) {
	runner := runtimeSelectionRunner{available: map[string]bool{chauftruntime.PHPImage("8.3"): true}}
	got := installedPHPForRuntimeWithRunner([]string{"7.4", "8.3"}, string(chauftruntime.EnginePodman), runner)
	if got["7.4"] {
		t.Fatal("native PHP state leaked into Podman selection")
	}
	if !got["8.3"] {
		t.Fatal("available Podman PHP image was not selected")
	}
}
