package commands

import (
	"fmt"

	"github.com/siegg/chauffeur/internal/lib"
	"github.com/siegg/chauffeur/internal/workspace"
)

func printRuntime(cfg workspace.Config) {
	lib.Pair("Runtime", cfg.Runtime.Engine)
}

func printRuntimeStep(cfg workspace.Config) {
	lib.Step(fmt.Sprintf("runtime:    %s", cfg.Runtime.Engine))
}
