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
	Nginx        NginxConfig
	PHP          PHPConfig
	Ports        PortConfig
	ProjectsDir  string
}

type NginxConfig struct {
	Enable    bool
	HTTPPort  int
	HTTPSPort int
}

type PHPConfig struct {
	Default       string            `yaml:"default"`
	LocalTarballs map[string]string `yaml:"local_tarballs"`
}

type PortConfig struct {
	// Port ranges for automatic allocation
	StartRange int `yaml:"start_range"`
	EndRange   int `yaml:"end_range"`

	// Port conflict resolution behavior
	ConflictResolution string `yaml:"conflict_resolution"` // "prompt", "auto", "fail"

	// Preferred ports for each service (fallback if not set in service configs)
	NginxHTTPFallback  int `yaml:"nginx_http_fallback"`
	NginxHTTPSFallback int `yaml:"nginx_https_fallback"`
	PHPFPMFallback     int `yaml:"php_fpm_fallback"`
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

func GetLocalTarballPath(version string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	if cfg.PHP.LocalTarballs == nil {
		return "", nil
	}

	return cfg.PHP.LocalTarballs[version], nil
}

func SetLocalTarballPath(version, path string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	if cfg.PHP.LocalTarballs == nil {
		cfg.PHP.LocalTarballs = make(map[string]string)
	}

	cfg.PHP.LocalTarballs[version] = path
	return Save(cfg)
}

func filePath() (string, error) {
	dir, err := workspace.Path(configDirName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func defaultConfig() (Config, error) {
	return DefaultConfig()
}

// DefaultConfig creates a default configuration with user-space ports
func DefaultConfig() (Config, error) {
	root, err := workspace.Dir()
	if err != nil {
		return Config{}, err
	}
	return Config{
		Version:      configVersion,
		Telemetry:    false,
		WorkspaceDir: root,
		Nginx: NginxConfig{
			Enable:    true,
			HTTPPort:  8080,
			HTTPSPort: 8443,
		},
		PHP:         PHPConfig{Default: "8.3"},
		ProjectsDir: filepath.Join(root, "projects"),
		Ports: PortConfig{
			StartRange:         8080,
			EndRange:           8099,
			ConflictResolution: "prompt",
			NginxHTTPFallback:  8080,
			NginxHTTPSFallback: 8443,
			PHPFPMFallback:     9000,
		},
	}, nil
}

func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = configVersion
	}
	c.WorkspaceDir = expandUserPath(c.WorkspaceDir)
	if c.WorkspaceDir == "" {
		if root, err := workspace.Dir(); err == nil {
			c.WorkspaceDir = root
		}
	}
	if c.ProjectsDir == "" {
		c.ProjectsDir = filepath.Join(c.WorkspaceDir, "projects")
	}
	c.ProjectsDir = expandUserPath(c.ProjectsDir)
	if !filepath.IsAbs(c.ProjectsDir) && c.WorkspaceDir != "" {
		c.ProjectsDir = filepath.Join(c.WorkspaceDir, c.ProjectsDir)
	}

	// Apply user-space port defaults to avoid system conflicts
	if c.Nginx.HTTPPort == 0 {
		c.Nginx.HTTPPort = 8080
	}
	if c.Nginx.HTTPSPort == 0 {
		c.Nginx.HTTPSPort = 8443
	}
	if c.PHP.Default == "" {
		c.PHP.Default = "8.3"
	}

	// Apply port management defaults
	if c.Ports.StartRange == 0 {
		c.Ports.StartRange = 8080
	}
	if c.Ports.EndRange == 0 {
		c.Ports.EndRange = 8099
	}
	if c.Ports.ConflictResolution == "" {
		c.Ports.ConflictResolution = "prompt"
	}
}

// ApplyDefaults exposes applyDefaults for external callers (tests/tools).
func (c *Config) ApplyDefaults() {
	c.applyDefaults()
}

func expandUserPath(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		if p == "~" {
			return home
		}
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(home, p[2:])
		}
	}
	return p
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
			case "nginx", "php":
				currentSection = key
			}
			continue
		}

		switch currentSection {
		case "nginx":
			switch key {
			case "enable":
				if v, err := strconv.ParseBool(value); err == nil {
					cfg.Nginx.Enable = v
				}
			case "http_port":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Nginx.HTTPPort = v
				}
			case "https_port":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Nginx.HTTPSPort = v
				}
			}
		case "php":
			if key == "default" && value != "" {
				cfg.PHP.Default = value
			} else if key == "local_tarballs" {
				// Initialize the map if nil
				if cfg.PHP.LocalTarballs == nil {
					cfg.PHP.LocalTarballs = make(map[string]string)
				}
			} else if strings.HasPrefix(key, "  ") && currentSection == "php" {
				// Handle indented tarball paths (e.g., "  8.3: /path/to/php-8.3.27.tar.gz")
				tarballKey := strings.TrimSpace(key)
				if tarballKey != "" && value != "" {
					if cfg.PHP.LocalTarballs == nil {
						cfg.PHP.LocalTarballs = make(map[string]string)
					}
					cfg.PHP.LocalTarballs[tarballKey] = value
				}
			}
		case "ports":
			switch key {
			case "start_range":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Ports.StartRange = v
				}
			case "end_range":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Ports.EndRange = v
				}
			case "conflict_resolution":
				if value != "" {
					cfg.Ports.ConflictResolution = value
				}
			case "nginx_http_fallback":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Ports.NginxHTTPFallback = v
				}
			case "nginx_https_fallback":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Ports.NginxHTTPSFallback = v
				}
			case "php_fpm_fallback":
				if v, err := strconv.Atoi(value); err == nil {
					cfg.Ports.PHPFPMFallback = v
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan config: %w", err)
	}
	return nil
}

func renderYAML(cfg Config) string {
	var localTarballsYAML string
	if len(cfg.PHP.LocalTarballs) > 0 {
		localTarballsYAML = "  local_tarballs:\n"
		for version, path := range cfg.PHP.LocalTarballs {
			localTarballsYAML += fmt.Sprintf("    %s: %s\n", version, path)
		}
	}

	return fmt.Sprintf(`version: %d
telemetry: %t
workspace_dir: %s
nginx:
  enable: %t
  http_port: %d
  https_port: %d
php:
  default: %s
%sports:
  start_range: %d
  end_range: %d
  conflict_resolution: %s
  nginx_http_fallback: %d
  nginx_https_fallback: %d
  php_fpm_fallback: %d
projects_dir: %s
`,
		cfg.Version,
		cfg.Telemetry,
		cfg.WorkspaceDir,
		cfg.Nginx.Enable,
		cfg.Nginx.HTTPPort,
		cfg.Nginx.HTTPSPort,
		cfg.PHP.Default,
		localTarballsYAML,
		cfg.Ports.StartRange,
		cfg.Ports.EndRange,
		cfg.Ports.ConflictResolution,
		cfg.Ports.NginxHTTPFallback,
		cfg.Ports.NginxHTTPSFallback,
		cfg.Ports.PHPFPMFallback,
		cfg.ProjectsDir,
	)
}
