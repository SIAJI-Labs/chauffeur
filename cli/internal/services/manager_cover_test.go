package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/siaji/chauffeur/cli/internal/projects"
)

func TestCreateVersionSpecificService(t *testing.T) {
	ws := t.TempDir()
	sm := &ServiceManager{workspaceDir: ws}

	phpDir := filepath.Join(ws, "php", "8.3")
	phpFpmBin := filepath.Join(phpDir, "sbin", "php-fpm")
	if err := os.MkdirAll(filepath.Dir(phpFpmBin), 0o755); err != nil {
		t.Fatalf("mkdirs: %v", err)
	}
	if err := os.WriteFile(phpFpmBin, []byte("fake"), 0o755); err != nil {
		t.Fatalf("write fake php-fpm: %v", err)
	}

	services, err := sm.createVersionSpecificService("8.3", phpFpmBin)
	if err != nil {
		t.Fatalf("createVersionSpecificService error: %v", err)
	}
	if len(services) != 1 || services[0].PIDFile == "" {
		t.Fatalf("expected one service with PID file, got %+v", services)
	}
	if _, err := os.Stat(filepath.Join(phpDir, "runtime", "php-fpm", "php-fpm.conf")); err != nil {
		t.Fatalf("expected config file: %v", err)
	}
}

func TestCreateProjectSpecificServiceAndSocketPath(t *testing.T) {
	ws := t.TempDir()
	projectsDir := filepath.Join(ws, "projects")
	sm := &ServiceManager{workspaceDir: ws, projectsDir: projectsDir}

	projectSlug := "demo"
	projectDir := filepath.Join(projectsDir, projectSlug)
	runtimeDir := filepath.Join(projectDir, "runtime", "php-fpm")
	logsDir := filepath.Join(projectDir, "logs")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatalf("mkdir runtime: %v", err)
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		t.Fatalf("mkdir logs: %v", err)
	}
	configPath := filepath.Join(projectDir, "project.yaml")
	cfg := projects.Config{
		Path: "/tmp/demo",
		PHP:  "8.3",
		Runtime: projects.Runtime{
			PHPFPM: filepath.Join(runtimeDir, "php-fpm.sock"),
			FPM:    &projects.FPM{Dedicated: true, Socket: filepath.Join(runtimeDir, "php-fpm.sock")},
		},
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := projects.WriteConfig(cfg, configPath, true); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	phpFpmBin := filepath.Join(ws, "php", "8.3", "bin", "php-fpm")
	if err := os.MkdirAll(filepath.Dir(phpFpmBin), 0o755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	if err := os.WriteFile(phpFpmBin, []byte("fake"), 0o755); err != nil {
		t.Fatalf("write fake php-fpm: %v", err)
	}

	services, err := sm.createProjectSpecificService(projectSlug, cfg, phpFpmBin)
	if err != nil {
		t.Fatalf("createProjectSpecificService error: %v", err)
	}
	if len(services) != 1 || services[0].Slug != projectSlug {
		t.Fatalf("unexpected services: %+v", services)
	}
	if _, err := os.Stat(filepath.Join(runtimeDir, "php-fpm.conf")); err != nil {
		t.Fatalf("expected project php-fpm.conf: %v", err)
	}

	socket, err := sm.GetSocketPath(projectSlug, "8.3")
	if err != nil {
		t.Fatalf("GetSocketPath error: %v", err)
	}
	if socket != filepath.Join(runtimeDir, "php-fpm.sock") {
		t.Fatalf("expected dedicated socket path, got %s", socket)
	}
}

func TestGetSocketPathSharedFallback(t *testing.T) {
	ws := t.TempDir()
	sm := &ServiceManager{workspaceDir: ws, projectsDir: filepath.Join(ws, "projects")}
	socket, err := sm.GetSocketPath("missing", "8.3")
	if err != nil {
		t.Fatalf("GetSocketPath error: %v", err)
	}
	expected := filepath.Join(ws, "php", "8.3", "runtime", "php-fpm", "php-fpm.sock")
	if socket != expected {
		t.Fatalf("expected shared socket %s, got %s", expected, socket)
	}
}

func TestIsRunningRemovesStalePid(t *testing.T) {
	ws := t.TempDir()
	pidFile := filepath.Join(ws, "service.pid")
	if err := os.WriteFile(pidFile, []byte("999999\n"), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}
	sm := &ServiceManager{}
	running, err := sm.IsRunning(Service{PIDFile: pidFile})
	if err != nil {
		t.Fatalf("IsRunning error: %v", err)
	}
	if running {
		t.Fatalf("expected not running")
	}
	if _, err := os.Stat(pidFile); err == nil {
		t.Fatalf("expected stale pid file to be removed")
	}
}
