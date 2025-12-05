package commands

import "github.com/siaji/chauffeur/cli/internal/services"

// OverrideServiceManager allows tests to inject fake service managers.
func OverrideServiceManager(fn func() (*services.ServiceManager, error)) (reset func()) {
	prev := newServiceManager
	newServiceManager = fn
	return func() { newServiceManager = prev }
}
