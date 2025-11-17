package projects

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	projectConfigName = "project.yaml"
	// ConfigVersion tracks the current project configuration schema version.
	ConfigVersion = 1
)

// ErrProjectNotFound is returned when a project configuration cannot be located.
var ErrProjectNotFound = errors.New("project not registered")

// Config represents the persisted project configuration schema.
type Config struct {
	Version   int
	Path      string
	PHP       string
	Site      *Site
	Runtime   Runtime
	CreatedAt time.Time
}

// Site holds optional site configuration for the project.
type Site struct {
	Domain string
	SSL    bool
}

// FPM holds PHP-FPM configuration for the project.
type FPM struct {
	Dedicated bool   `yaml:"dedicated"`
	Socket    string `yaml:"socket"`
}

// Runtime tracks runtime layout paths for the project.
type Runtime struct {
	PHPFPM string
	FPM    *FPM `yaml:"fpm"`
}

// Layout captures important directories for a project slug.
type Layout struct {
	Root       string
	ConfigPath string
	RuntimeDir string
	LogsDir    string
	SocketPath string
}

// EnsureLayout prepares the directory structure for a project slug under the provided base directory.
func EnsureLayout(baseDir, slug string) (Layout, error) {
	if baseDir == "" {
		return Layout{}, fmt.Errorf("projects base directory is empty")
	}

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return Layout{}, fmt.Errorf("ensure projects directory: %w", err)
	}

	root := filepath.Join(baseDir, slug)
	runtimeDir := filepath.Join(root, "runtime", "php-fpm")
	logsDir := filepath.Join(root, "logs")

	for _, dir := range []string{root, runtimeDir, logsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Layout{}, fmt.Errorf("ensure project directory %s: %w", dir, err)
		}
	}

	layout := Layout{
		Root:       root,
		ConfigPath: filepath.Join(root, projectConfigName),
		RuntimeDir: runtimeDir,
		LogsDir:    logsDir,
		SocketPath: filepath.Join(runtimeDir, "php-fpm.sock"),
	}

	return layout, nil
}

// WriteConfig renders and writes a project configuration file.
func WriteConfig(cfg Config, path string, force bool) error {
	if path == "" {
		return fmt.Errorf("project config path is empty")
	}

	if cfg.Version == 0 {
		cfg.Version = ConfigVersion
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now().UTC()
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("project configuration already exists at %s (use --force to overwrite)", path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check existing project config: %w", err)
		}
	}

	data := renderYAML(cfg)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return fmt.Errorf("write project config: %w", err)
	}
	return nil
}

// LoadConfig loads and parses a project configuration file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read project config: %w", err)
	}
	cfg, err := parseConfig(data)
	if err != nil {
		return Config{}, fmt.Errorf("parse project config: %w", err)
	}
	return cfg, nil
}

// FindByPath locates the project configuration whose path matches projectPath.
func FindByPath(baseDir, projectPath string) (Config, Layout, error) {
	if baseDir == "" {
		return Config{}, Layout{}, fmt.Errorf("projects base directory is empty")
	}

	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return Config{}, Layout{}, fmt.Errorf("resolve project path: %w", err)
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, Layout{}, fmt.Errorf("%w: %s", ErrProjectNotFound, absProjectPath)
		}
		return Config{}, Layout{}, fmt.Errorf("read projects directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		layout, err := EnsureLayout(baseDir, entry.Name())
		if err != nil {
			return Config{}, Layout{}, err
		}
		cfg, err := LoadConfig(layout.ConfigPath)
		if err != nil {
			return Config{}, Layout{}, fmt.Errorf("load project config %s: %w", layout.ConfigPath, err)
		}

		storedPath := cfg.Path
		if absStored, err := filepath.Abs(cfg.Path); err == nil {
			storedPath = absStored
		}
		if samePaths(storedPath, absProjectPath) {
			if cfg.Runtime.PHPFPM == "" {
				cfg.Runtime.PHPFPM = layout.SocketPath
			}
			return cfg, layout, nil
		}
	}

	return Config{}, Layout{}, fmt.Errorf("%w: %s", ErrProjectNotFound, absProjectPath)
}

// Slugify converts a directory name into a safe project slug.
func Slugify(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "project"
	}

	var b strings.Builder
	b.Grow(len(name))

	lastWasHyphen := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastWasHyphen = false
			continue
		}
		if !lastWasHyphen {
			b.WriteByte('-')
			lastWasHyphen = true
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "project"
	}
	return slug
}

func renderYAML(cfg Config) string {
	version := cfg.Version
	if version == 0 {
		version = ConfigVersion
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("version: %d\n", version))
	b.WriteString(fmt.Sprintf("path: %s\n", cfg.Path))
	b.WriteString(fmt.Sprintf("php: %s\n", cfg.PHP))

	if cfg.Site != nil {
		b.WriteString("site:\n")
		if cfg.Site.Domain != "" {
			b.WriteString(fmt.Sprintf("  domain: %s\n", cfg.Site.Domain))
		}
		b.WriteString(fmt.Sprintf("  ssl: %t\n", cfg.Site.SSL))
	}

	b.WriteString("runtime:\n")
	b.WriteString(fmt.Sprintf("  php_fpm_socket: %s\n", cfg.Runtime.PHPFPM))

	if cfg.Runtime.FPM != nil {
		b.WriteString("  fpm:\n")
		b.WriteString(fmt.Sprintf("    dedicated: %t\n", cfg.Runtime.FPM.Dedicated))
		if cfg.Runtime.FPM.Socket != "" {
			b.WriteString(fmt.Sprintf("    socket: %s\n", cfg.Runtime.FPM.Socket))
		}
	}

	b.WriteString(fmt.Sprintf("created_at: %s\n", cfg.CreatedAt.UTC().Format(time.RFC3339)))
	b.WriteString("")

	return b.String()
}

func parseConfig(data []byte) (Config, error) {
	var cfg Config
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
			case "path":
				cfg.Path = value
			case "php":
				cfg.PHP = value
			case "created_at":
				if value != "" {
					if ts, err := time.Parse(time.RFC3339, value); err == nil {
						cfg.CreatedAt = ts
					}
				}
			case "site", "runtime":
				currentSection = key
				if key == "site" && cfg.Site == nil {
					cfg.Site = &Site{}
				}
			}
			continue
		}

		switch currentSection {
		case "site":
			switch key {
			case "domain":
				cfg.Site.Domain = value
			case "ssl":
				if v, err := strconv.ParseBool(value); err == nil {
					cfg.Site.SSL = v
				}
			}
		case "runtime":
			if key == "php_fpm_socket" {
				cfg.Runtime.PHPFPM = value
			} else if key == "fpm" {
				// Initialize FPM config if nil
				if cfg.Runtime.FPM == nil {
					cfg.Runtime.FPM = &FPM{}
				}
			} else if cfg.Runtime.FPM != nil {
				if key == "dedicated" {
					if v, err := strconv.ParseBool(value); err == nil {
						cfg.Runtime.FPM.Dedicated = v
					}
				} else if key == "socket" {
					cfg.Runtime.FPM.Socket = value
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("scan project config: %w", err)
	}

	return cfg, nil
}

func samePaths(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

// IsPHPVersionInstalled checks if a PHP version is installed in the workspace
func IsPHPVersionInstalled(version string) bool {
	// Get the workspace directory
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	workspaceDir := filepath.Join(home, ".chauffeur")
	
	// Check if PHP binary exists
	binary := filepath.Join(workspaceDir, "php", version, "bin", "php")
	if _, err := os.Stat(binary); err != nil {
		return false
	}
	return true
}
