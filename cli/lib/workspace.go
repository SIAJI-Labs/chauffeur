package lib

import (
	"fmt"

	"github.com/siaji/chauffeur/cli/internal/workspace"
)

// ValidateWorkspace validates that the Chauffeur workspace exists and offers to initialize it if needed.
// This is a reusable helper for commands that require a workspace to function.
// Returns an error if validation fails or if workspace initialization is required and user declines.
//
// Usage:
//
//	if err := lib.ValidateWorkspace(args); err != nil {
//	    return err
//	}
//
// Args should be the command arguments passed to the command function.
// Help flags (--help, -h) are automatically skipped to allow help text to be displayed without workspace validation.
func ValidateWorkspace(args []string) error {
	logger := NewCommandLogger("workspace")
	// Check if workspace is ready (skips validation for help commands)
	if ready, err := workspace.ValidateForCommandWithLogger(args, logger); err != nil {
		return fmt.Errorf("workspace validation failed: %w", err)
	} else if !ready {
		return fmt.Errorf("workspace initialization required")
	}

	return nil
}
