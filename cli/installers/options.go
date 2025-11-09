package installers

import (
	"net/http"

	"github.com/siaji/chauffeur/cli/internal/system"
)

// InstallOptions provides shared configuration for installer functions.
type InstallOptions struct {
	Prefix string       // Chauffeur workspace prefix (e.g., ~/.chauffeur)
	Force  bool         // Reinstall even if binaries already exist
	Client *http.Client // Optional HTTP client (defaults applied when nil)
	Info   system.Info  // Host system metadata (distro/arch) for logging/build decisions
}
