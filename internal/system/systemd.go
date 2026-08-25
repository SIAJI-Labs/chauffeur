package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	nginxUnit       = "chauffeur-nginx.service"
	fpmTemplateUnit = "chauffeur-php-fpm@.service"
)

// UserSystemdDir returns the path to the user's systemd unit directory.
func UserSystemdDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user")
}

// NginxUnitContent returns the systemd user service unit for Chauffeur's nginx.
// Uses %h (systemd home specifier) so the unit is portable across users.
func NginxUnitContent() string {
	return `[Unit]
Description=Chauffeur nginx
After=network.target

[Service]
Type=forking
PIDFile=%h/.chauffeur/nginx/logs/nginx.pid
ExecStart=%h/.chauffeur/nginx/sbin/nginx -c %h/.chauffeur/nginx/etc/nginx.conf
ExecReload=/bin/kill -HUP $MAINPID
ExecStop=/bin/kill -QUIT $MAINPID
PrivateTmp=no
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
}

// FPMTemplateUnitContent returns the systemd user template service unit for
// Chauffeur PHP-FPM. %i is replaced with the PHP version when instantiated
// (e.g. chauffeur-php-fpm@8.3.service).
func FPMTemplateUnitContent() string {
	return `[Unit]
Description=Chauffeur PHP-FPM (%i)
After=chauffeur-nginx.service

[Service]
Type=forking
PIDFile=%h/.chauffeur/php/%i/runtime/php-fpm/php-fpm.pid
ExecStart=%h/.chauffeur/php/%i/sbin/php-fpm --fpm-config %h/.chauffeur/php/%i/etc/php-fpm.conf --daemonize
ExecReload=/bin/kill -USR2 $MAINPID
ExecStop=/bin/kill -TERM $MAINPID
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`
}

// FPMInstanceUnit returns the systemd unit name for a specific PHP version.
func FPMInstanceUnit(version string) string {
	return fmt.Sprintf("chauffeur-php-fpm@%s.service", version)
}

// WriteNginxUnit writes the nginx systemd user unit file.
func WriteNginxUnit() error {
	return writeUnit(nginxUnit, NginxUnitContent())
}

// WriteFPMTemplateUnit writes the PHP-FPM template systemd user unit file.
func WriteFPMTemplateUnit() error {
	return writeUnit(fpmTemplateUnit, FPMTemplateUnitContent())
}

// UserDaemonReload runs systemctl --user daemon-reload.
func UserDaemonReload() error {
	return userctl("daemon-reload")
}

// EnableUnit runs systemctl --user enable <unit> (persistence only, no start).
func EnableUnit(unit string) error {
	return userctl("enable", unit)
}

// StartUnit runs systemctl --user start <unit>.
func StartUnit(unit string) error {
	return userctl("start", unit)
}

// StopUnit runs systemctl --user stop <unit> without disabling auto-start.
func StopUnit(unit string) error {
	return userctl("stop", unit)
}

// DisableUnit runs systemctl --user disable --now <unit>.
func DisableUnit(unit string) error {
	return userctl("disable", "--now", unit)
}

// IsUnitActive returns true if the given user unit is currently active.
func IsUnitActive(unit string) bool {
	out, err := exec.Command("systemctl", "--user", "is-active", unit).Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

// IsUnitEnabled returns true if the given user unit is enabled.
func IsUnitEnabled(unit string) bool {
	out, err := exec.Command("systemctl", "--user", "is-enabled", unit).Output()
	return err == nil && strings.TrimSpace(string(out)) == "enabled"
}

// IsLingeringEnabled returns true if systemd lingering is active for the
// current user (required for services to start on boot, not just on login).
func IsLingeringEnabled() bool {
	user := os.Getenv("USER")
	if user == "" {
		return false
	}
	out, err := exec.Command("loginctl", "show-user", user,
		"--property=Linger", "--value").Output()
	return err == nil && strings.TrimSpace(string(out)) == "yes"
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeUnit(name, content string) error {
	dir := UserSystemdDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create systemd user dir: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
}

func userctl(args ...string) error {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("systemctl --user %s: %s", args[0], msg)
		}
		return fmt.Errorf("systemctl --user %s: %w", args[0], err)
	}
	return nil
}
