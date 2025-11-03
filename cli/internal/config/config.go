package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/siaji/chauffeur/cli/internal/workspace"
)

const (
	configDirName  = "config"
	configFileName = "chauffeur.yaml"
	configVersion  = 1
)

type Config struct {
	Version      int
	Telemetry    bool
	WorkspaceDir string
	Caddy        CaddyConfig
	Nginx        NginxConfig
	PHP          PHPConfig
	ProjectsDir  string
}

type CaddyConfig struct {
	Enable    bool
	HTTPPort  int
	HTTPSPort int
}

type NginxConfig struct {
	Enable bool
}

type PHPConfig struct {
	Default string
}

func Load() (Config, error) {
	path, err := filePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultConfig()
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, err
	}

	if err := parseIntoConfig(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.applyDefaults()

	return cfg, nil
}

func Save(cfg Config) error {
	cfg.applyDefaults()

	dir, err := workspace.Path(configDirName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	if err := os.WriteFile(path, []byte(renderYAML(cfg)), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func SetDefaultPHPVersion(version string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.PHP.Default = version
	return Save(cfg)
}

func GetDefaultPHPVersion() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	return cfg.PHP.Default, nil
}

func filePath() (string, error) {
	dir, err := workspace.Path(configDirName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func defaultConfig() (Config, error) {
	root, err := workspace.Dir()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Version:      configVersion,
		Telemetry:    false,
		WorkspaceDir: root,
		Caddy: CaddyConfig{
			Enable:    true,
			HTTPPort:  80,
			HTTPSPort: 443,
		},
		Nginx:       NginxConfig{Enable: true},
		PHP:         PHPConfig{Default: "8.3"},
		ProjectsDir: filepath.Join(root, "projects"),
	}, nil
}

func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = configVersion
	}
	if c.WorkspaceDir == "" {
		if root, err := workspace.Dir(); err == nil {
			c.WorkspaceDir = root
		}
	}
	if c.ProjectsDir == "" {
		c.ProjectsDir = filepath.Join(c.WorkspaceDir, "projects")
	}
	if c.Caddy.HTTPPort == 0 {
		c.Caddy.HTTPPort = 80
	}
	if c.Caddy.HTTPSPort == 0 {
		c.Caddy.HTTPSPort = 443
	}
	if c.PHP.Default == "" {
		c.PHP.Default = "8.3"
	}
}

func parseIntoConfig(data []byte, cfg *Config) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			currentSection = ""
		}

		if !strings.Contains(trimmed, ":") {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
		}

		if indent == 0 {
			switch key {
			case "version":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Version = v
				}
			case "telemetry":
				if v, err := strconv.ParseBool(value); err == nil {
					cfg.Telemetry = v
				}
			case "workspace_dir":
				if value != "" {
					cfg.WorkspaceDir = value
				}
			case "projects_dir":
				if value != "" {
					cfg.ProjectsDir = value
				}
			case "caddy", "nginx", "php":
				currentSection = key
			}
			continue
		}

		switch currentSection {
		case "caddy":
			switch key {
			case "enable":
				if v, err := strconv.ParseBool(value); err == nil {
					cfg.Caddy.Enable = v
				}
			case "http_port":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Caddy.HTTPPort = v
				}
			case "https_port":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Caddy.HTTPSPort = v
				}
			}
		case "nginx":
			if key == "enable" {
				if v, err := strconv.ParseBool(value); err == nil {
					cfg.Nginx.Enable = v
				}
			}
		case "php":
			if key == "default" && value != "" {
				cfg.PHP.Default = value
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan config: %w", err)
	}
	return nil
}

func renderYAML(cfg Config) string {
	return fmt.Sprintf("version: %d\ntelemetry: %t\nworkspace_dir: %s\ncaddy:\n  enable: %t\n  http_port: %d\n  https_port: %d\nnginx:\n  enable: %t\nphp:\n  default: %s\nprojects_dir: %s\n",
		cfg.Version,
		cfg.Telemetry,
		cfg.WorkspaceDir,
		cfg.Caddy.Enable,
		cfg.Caddy.HTTPPort,
		cfg.Caddy.HTTPSPort,
		cfg.Nginx.Enable,
		cfg.PHP.Default,
		cfg.ProjectsDir,
	)
}
