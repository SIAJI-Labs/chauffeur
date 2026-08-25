package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/system"
	"github.com/siegg/chauffeur/internal/workspace"
)

const stopTimeout = 30 * time.Second

// Manager orchestrates start/stop/restart of all Chauffeur services.
type Manager struct {
	root string
}

// NewManager returns a Manager for the given workspace root.
func NewManager(root string) *Manager {
	return &Manager{root: root}
}

// Nginx returns the NginxService for this workspace.
func (m *Manager) Nginx() *NginxService {
	return NewNginxService(m.root)
}

// SharedFPM returns the shared FPMService for a PHP version.
func (m *Manager) SharedFPM(version string) *FPMService {
	return NewSharedFPM(m.root, version)
}

// AllFPM returns all FPM services needed by currently linked projects:
// one shared pool per unique PHP version + one dedicated pool per project
// with fpm.dedicated: true. Shared pools are returned first, sorted by version.
func (m *Manager) AllFPM() ([]*FPMService, error) {
	all, err := projects.ListAll(m.root)
	if err != nil {
		return nil, err
	}

	sharedVersions := map[string]bool{}
	var dedicated []*FPMService

	for _, p := range all {
		if p.ProjectType == projects.TypeReverseProxy || p.PHPVersion == "" {
			continue
		}
		if p.FPM.Dedicated {
			dedicated = append(dedicated, NewDedicatedFPM(m.root, p.Slug, p.PHPVersion, p.FPM.Socket))
		} else {
			sharedVersions[p.PHPVersion] = true
		}
	}

	versions := make([]string, 0, len(sharedVersions))
	for v := range sharedVersions {
		versions = append(versions, v)
	}
	sort.Strings(versions)

	var fpms []*FPMService
	for _, v := range versions {
		fpms = append(fpms, NewSharedFPM(m.root, v))
	}
	return append(fpms, dedicated...), nil
}

// StartAll starts all services in the correct order:
// shared FPM → dedicated FPM → nginx.
func (m *Manager) StartAll() ([]StartResult, error) {
	nginx := m.Nginx()
	fpms, err := m.AllFPM()
	if err != nil {
		return nil, err
	}
	cfg := workspace.Load()

	var results []StartResult

	// 1. Start FPM pools first
	for _, fpm := range fpms {
		r := StartResult{Label: "php-fpm " + fpm.Label()}
		if fpm.IsRunning() {
			r.AlreadyRunning = true
			r.PID = fpm.PID()
		} else {
			if err := fpm.Start(); err != nil {
				r.Err = err
			} else {
				r.PID = fpm.PID()
				// Wait for socket before nginx starts routing to it
				_ = waitForSocket(fpm.SockPath(), 5*time.Second)
			}
		}
		results = append(results, r)
	}

	// 2. Check port availability before starting nginx
	if !nginx.IsRunning() {
		for _, port := range []int{cfg.Nginx.HTTPPort, cfg.Nginx.HTTPSPort} {
			if !IsPortAvailable(port) {
				pid, name, _ := FindProcessOnPort(port)
				msg := fmt.Sprintf("port %d is already in use", port)
				if pid > 0 {
					msg += fmt.Sprintf(" by %s (pid %d)", name, pid)
				}
				return results, fmt.Errorf("%s\n\n  Kill it:\n    kill %d\n\n  Or change port:\n    chauf config set nginx.http_port <port>", msg, pid)
			}
		}
	}

	// 3. Start nginx last
	r := StartResult{Label: "nginx"}
	if nginx.IsRunning() {
		r.AlreadyRunning = true
		r.PID = nginx.PID()
	} else {
		if err := nginx.Start(); err != nil {
			r.Err = err
		} else {
			r.PID = nginx.PID()
		}
	}
	results = append(results, r)

	return results, nil
}

// StopAll stops all services in reverse order: nginx → FPM.
func (m *Manager) StopAll() ([]StopResult, error) {
	nginx := m.Nginx()
	fpms, err := m.AllFPM()
	if err != nil {
		return nil, err
	}

	var results []StopResult

	// 1. Stop nginx first (stop accepting new requests)
	r := StopResult{Label: "nginx"}
	if system.IsUnitActive("chauffeur-nginx.service") {
		r.Err = system.StopUnit("chauffeur-nginx.service")
	} else if !nginx.IsRunning() {
		r.AlreadyStopped = true
	} else {
		r.Err = nginx.Stop(stopTimeout)
	}
	results = append(results, r)

	// 2. Stop FPM pools
	for _, fpm := range fpms {
		r := StopResult{Label: "php-fpm " + fpm.Label()}
		unit := system.FPMInstanceUnit(fpm.Version())
		if fpm.Version() != "" && system.IsUnitActive(unit) {
			r.Err = system.StopUnit(unit)
		} else if !fpm.IsRunning() {
			r.AlreadyStopped = true
		} else {
			r.Err = fpm.Stop(stopTimeout)
		}
		results = append(results, r)
	}

	return results, nil
}

// ReloadNginx does a zero-downtime nginx config reload via SIGHUP.
func (m *Manager) ReloadNginx() error {
	return m.Nginx().Reload()
}

// ReloadFPM sends SIGUSR2 to the shared FPM pool for a specific PHP version.
func (m *Manager) ReloadFPM(version string) error {
	return NewSharedFPM(m.root, version).Reload()
}

// ReloadAll reloads nginx and all running FPM pools gracefully.
func (m *Manager) ReloadAll() error {
	fpms, err := m.AllFPM()
	if err != nil {
		return err
	}
	for _, fpm := range fpms {
		if fpm.IsRunning() {
			_ = fpm.Reload()
		}
	}
	return m.Nginx().Reload()
}

// ── Result types ──────────────────────────────────────────────────────────────

// StartResult carries the outcome of starting a single service.
type StartResult struct {
	Label          string
	PID            int
	AlreadyRunning bool
	Err            error
}

// StopResult carries the outcome of stopping a single service.
type StopResult struct {
	Label          string
	AlreadyStopped bool
	Err            error
}
