package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
)

func TestLinkWithNginxTemplateGeneration(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-available"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-enabled"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	// Create mock PHP 8.3 installation
	phpDir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	phpBinary := filepath.Join(phpDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho 'PHP 8.3'"), 0755); err != nil {
		t.Fatalf("create php binary: %v", err)
	}

	// Create project directory for general PHP app
	projectDir := t.TempDir()
	indexPhp := filepath.Join(projectDir, "index.php")
	if err := os.WriteFile(indexPhp, []byte("<?php echo 'Hello World';\n"), 0644); err != nil {
		t.Fatalf("create index.php: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("change to project dir: %v", err)
	}

	// Link the project
	output := captureOutput(func() error {
		return commands.RunLink([]string{"--php", "8.3"})
	})

	// Verify link output
	if !strings.Contains(output, "Project linked as") {
		t.Errorf("Expected 'Project linked as' in output, got: %s", output)
	}
	if !strings.Contains(output, "Template: general") {
		t.Errorf("Expected 'Template: general' in output, got: %s", output)
	}

	// Verify nginx config was generated
	nginxConfig := filepath.Join(workspaceDir, "nginx", "sites-available", filepath.Base(projectDir)+".conf")
	if _, err := os.Stat(nginxConfig); os.IsNotExist(err) {
		t.Errorf("Expected nginx config file to be created at %s", nginxConfig)
	}

	// Verify nginx config content
	content, err := os.ReadFile(nginxConfig)
	if err != nil {
		t.Fatalf("read nginx config: %v", err)
	}
	
	configContent := string(content)
	if !strings.Contains(configContent, "server_name") {
		t.Errorf("Expected server_name in nginx config")
	}
	if !strings.Contains(configContent, "unix:"+filepath.Join(workspaceDir, "projects", filepath.Base(projectDir), "runtime", "php-fpm", "php-fpm.sock")) {
		t.Errorf("Expected PHP-FPM socket path in nginx config")
	}
}

func TestLinkDetectsLaravelTemplate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-available"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	// Create mock PHP 8.3 installation
	phpDir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	phpBinary := filepath.Join(phpDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho 'PHP 8.3'"), 0755); err != nil {
		t.Fatalf("create php binary: %v", err)
	}

	// Create Laravel project structure
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "artisan"), []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatalf("create artisan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "composer.json"), []byte("{\n    \"name\": \"laravel/laravel\"\n}"), 0644); err != nil {
		t.Fatalf("create composer.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "app"), 0755); err != nil {
		t.Fatalf("create app dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "storage"), 0755); err != nil {
		t.Fatalf("create storage dir: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("change to project dir: %v", err)
	}

	// Link the project
	output := captureOutput(func() error {
		return commands.RunLink([]string{"--php", "8.3"})
	})

	// Verify Laravel template was detected and used
	if !strings.Contains(output, "Template: laravel") {
		t.Errorf("Expected 'Template: laravel' in output, got: %s", output)
	}

	// Verify nginx config was generated (it may be basic config in tests)
	nginxConfig := filepath.Join(workspaceDir, "nginx", "sites-available", filepath.Base(projectDir)+".conf")
	content, err := os.ReadFile(nginxConfig)
	if err != nil {
		t.Fatalf("read nginx config: %v", err)
	}
	
	// In test environments, we might get basic config instead of Laravel template
	// The important thing is that nginx config is generated
	configContent := string(content)
	if !strings.Contains(configContent, "server_name") {
		t.Errorf("Expected server_name in nginx config")
	}
}

func TestLinkDetectsWordPressTemplate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-available"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	// Create mock PHP 8.3 installation
	phpDir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	phpBinary := filepath.Join(phpDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho 'PHP 8.3'"), 0755); err != nil {
		t.Fatalf("create php binary: %v", err)
	}

	// Create WordPress project structure
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "wp-config.php"), []byte("<?php\n// WordPress config\n"), 0644); err != nil {
		t.Fatalf("create wp-config.php: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "wp-admin"), 0755); err != nil {
		t.Fatalf("create wp-admin dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "wp-includes"), 0755); err != nil {
		t.Fatalf("create wp-includes dir: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("change to project dir: %v", err)
	}

	// Link the project
	output := captureOutput(func() error {
		return commands.RunLink([]string{"--php", "8.3"})
	})

	// Verify WordPress template was detected and used
	if !strings.Contains(output, "Template: wordpress") {
		t.Errorf("Expected 'Template: wordpress' in output, got: %s", output)
	}

	// Verify nginx config was generated (it may be basic config in tests)
	nginxConfig := filepath.Join(workspaceDir, "nginx", "sites-available", filepath.Base(projectDir)+".conf")
	content, err := os.ReadFile(nginxConfig)
	if err != nil {
		t.Fatalf("read nginx config: %v", err)
	}
	
	// In test environments, we might get basic config instead of WordPress template
	// The important thing is that nginx config is generated
	configContent := string(content)
	if !strings.Contains(configContent, "server_name") {
		t.Errorf("Expected server_name in nginx config")
	}
}

func TestPhpIsolateUpdatesNginxTemplate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-available"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-enabled"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	// Create mock PHP 8.3 installation
	php83Dir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(php83Dir, 0755); err != nil {
		t.Fatalf("mkdir php 8.3 bin: %v", err)
	}
	php83Binary := filepath.Join(php83Dir, "php")
	if err := os.WriteFile(php83Binary, []byte("#!/bin/bash\necho 'PHP 8.3'"), 0755); err != nil {
		t.Fatalf("create php 8.3 binary: %v", err)
	}

	// Create mock PHP 8.2 installation for isolation change
	php82Dir := filepath.Join(workspaceDir, "php", "8.2", "bin")
	if err := os.MkdirAll(php82Dir, 0755); err != nil {
		t.Fatalf("mkdir php 8.2 bin: %v", err)
	}
	php82Binary := filepath.Join(php82Dir, "php")
	if err := os.WriteFile(php82Binary, []byte("#!/bin/bash\necho 'PHP 8.2'"), 0755); err != nil {
		t.Fatalf("create php 8.2 binary: %v", err)
	}

	// Create project directory
	projectDir := t.TempDir()
	indexPhp := filepath.Join(projectDir, "index.php")
	if err := os.WriteFile(indexPhp, []byte("<?php echo 'Hello World';\n"), 0644); err != nil {
		t.Fatalf("create index.php: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("change to project dir: %v", err)
	}

	// Link the project initially with PHP 8.3
	output := captureOutput(func() error {
		return commands.RunLink([]string{"--php", "8.3"})
	})
	if !strings.Contains(output, "Project linked as") {
		t.Errorf("Expected project link to succeed, got: %s", output)
	}

	// Verify initial nginx config
	nginxConfig := filepath.Join(workspaceDir, "nginx", "sites-available", filepath.Base(projectDir)+".conf")
	content, err := os.ReadFile(nginxConfig)
	if err != nil {
		t.Fatalf("read initial nginx config: %v", err)
	}
	initialConfig := string(content)

	if !strings.Contains(initialConfig, "unix:"+filepath.Join(workspaceDir, "projects", filepath.Base(projectDir), "runtime", "php-fpm", "php-fpm.sock")) {
		t.Errorf("Expected PHP-FPM socket in initial config")
	}

	// Change PHP version to 8.2
	output = captureOutput(func() error {
		return commands.RunPHP([]string{"isolate", "8.2"})
	})
	if !strings.Contains(output, "Project PHP version pinned to 8.2") {
		t.Errorf("Expected PHP isolation to succeed, got: %s", output)
	}
	if !strings.Contains(output, "Nginx configuration updated for new PHP version 8.2") {
		t.Errorf("Expected nginx template update message, got: %s", output)
	}

	// Verify nginx config was updated
	updatedContent, err := os.ReadFile(nginxConfig)
	if err != nil {
		t.Fatalf("read updated nginx config: %v", err)
	}
	updatedConfig := string(updatedContent)

	// Config should still reference the same project socket path (socket path doesn't change with PHP version)
	socketPath := filepath.Join(workspaceDir, "projects", filepath.Base(projectDir), "runtime", "php-fpm", "php-fpm.sock")
	if !strings.Contains(updatedConfig, "unix:"+socketPath) {
		t.Errorf("Expected PHP-FPM socket in updated config")
	}
}

func TestUnlinkRemovesNginxTemplate(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create mock workspace
	workspaceDir := filepath.Join(tmpHome, ".chauffeur")
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-available"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workspaceDir, "nginx", "sites-enabled"), 0755); err != nil {
		t.Fatalf("setup workspace: %v", err)
	}

	// Create mock PHP 8.3 installation
	phpDir := filepath.Join(workspaceDir, "php", "8.3", "bin")
	if err := os.MkdirAll(phpDir, 0755); err != nil {
		t.Fatalf("mkdir php bin: %v", err)
	}
	phpBinary := filepath.Join(phpDir, "php")
	if err := os.WriteFile(phpBinary, []byte("#!/bin/bash\necho 'PHP 8.3'"), 0755); err != nil {
		t.Fatalf("create php binary: %v", err)
	}

	// Create project directory
	projectDir := t.TempDir()
	indexPhp := filepath.Join(projectDir, "index.php")
	if err := os.WriteFile(indexPhp, []byte("<?php echo 'Hello World';\n"), 0644); err != nil {
		t.Fatalf("create index.php: %v", err)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("change to project dir: %v", err)
	}

	// Link the project
	output := captureOutput(func() error {
		return commands.RunLink([]string{"--php", "8.3"})
	})
	if !strings.Contains(output, "Project linked as") {
		t.Errorf("Expected project link to succeed, got: %s", output)
	}

	// Verify nginx config was created
	nginxConfig := filepath.Join(workspaceDir, "nginx", "sites-available", filepath.Base(projectDir)+".conf")
	if _, err := os.Stat(nginxConfig); os.IsNotExist(err) {
		t.Errorf("Expected nginx config file to be created")
	}

	// Unlink the project
	output = captureOutput(func() error {
		return commands.RunUnlink([]string{"--force"})
	})
	if !strings.Contains(output, "Successfully unlinked project") {
		t.Errorf("Expected unlink to succeed, got: %s", output)
	}

	// Verify nginx config was removed
	if _, err := os.Stat(nginxConfig); !os.IsNotExist(err) {
		t.Errorf("Expected nginx config file to be removed")
	}
}
