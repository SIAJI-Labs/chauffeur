package runtime

import (
	"fmt"

	"github.com/siegg/chauffeur/internal/workspace"
)

// ForWorkspace returns the configured runtime. Native is the safe default
// until an explicit Podman migration is selected.
func ForWorkspace(cfg workspace.Config) (Runtime, error) {
	switch Engine(cfg.Runtime.Engine) {
	case "", EngineNative:
		return Native{Root: workspace.Root()}, nil
	case EnginePodman:
		return Podman{Runner: ExecRunner{}}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime engine %q", cfg.Runtime.Engine)
	}
}
