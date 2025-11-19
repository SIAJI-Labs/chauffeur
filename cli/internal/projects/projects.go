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
	"gopkg.in/yaml.v2"
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
	Domains   *Domains `yaml:"domains,omitempty"`
	Runtime   Runtime
	CreatedAt time.Time
}

// Site holds optional site configuration for the project.
type Site struct {
	Domain string
	SSL    bool
}

// DomainAlias represents an alias domain configuration.
type DomainAlias struct {
	Domain string `yaml:"domain"`
	SSL    bool   `yaml:"ssl"`
}

// Domains holds multi-domain configuration for the project.
type Domains struct {
	Aliases []DomainAlias `yaml:"aliases,omitempty"`
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

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal project config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
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
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
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

	if cfg.Domains != nil && len(cfg.Domains.Aliases) > 0 {
		b.WriteString("domains:\n")
		b.WriteString("  aliases:\n")
		for _, alias := range cfg.Domains.Aliases {
			b.WriteString(fmt.Sprintf("    - domain: %s\n", alias.Domain))
			b.WriteString(fmt.Sprintf("      ssl: %t\n", alias.SSL))
		}
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
		case "domains":
			// Just track that we're in the domains section - actual parsing handled by line detection
		case "aliases":
			// Aliases section - look for domain entries
			if strings.Contains(trimmed, "- domain:") {
				if cfg.Domains == nil {
					cfg.Domains = &Domains{}
				}
				// Extract domain from "- domain: cms-hja.test"
				if parts := strings.SplitN(trimmed, ":", 2); len(parts) == 2 {
					domain := strings.TrimSpace(parts[1])
					alias := DomainAlias{Domain: domain, SSL: false} // Default SSL to false
					cfg.Domains.Aliases = append(cfg.Domains.Aliases, alias)
				}
			} else if key == "ssl" && len(cfg.Domains.Aliases) > 0 {
				// Set SSL for the last added alias
				if v, err := strconv.ParseBool(value); err == nil {
					lastIndex := len(cfg.Domains.Aliases) - 1
					cfg.Domains.Aliases[lastIndex].SSL = v
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

// GetAllDomains returns all domains for a project (primary + aliases)
func (c *Config) GetAllDomains() []DomainAlias {
	var domains []DomainAlias

	// Add primary domain from site (legacy support)
	if c.Site != nil && c.Site.Domain != "" {
		domains = append(domains, DomainAlias{
			Domain: c.Site.Domain,
			SSL:    c.Site.SSL,
		})
	}

	// Add alias domains
	if c.Domains != nil {
		for _, alias := range c.Domains.Aliases {
			domains = append(domains, alias)
		}
	}

	return domains
}

// GetPrimaryDomain returns the primary domain for the project
func (c *Config) GetPrimaryDomain() string {
	if c.Site != nil && c.Site.Domain != "" {
		return c.Site.Domain
	}
	return ""
}

// GetServerNames returns a space-separated list of all server names for nginx config
func (c *Config) GetServerNames() string {
	domains := c.GetAllDomains()
	if len(domains) == 0 {
		return ""
	}

	var serverNames []string
	for _, domain := range domains {
		serverNames = append(serverNames, domain.Domain)
	}

	return strings.Join(serverNames, " ")
}

// HasSSLEnabled returns true if any domain (primary or alias) has SSL enabled
func (c *Config) HasSSLEnabled() bool {
	// Check primary domain
	if c.Site != nil && c.Site.SSL {
		return true
	}

	// Check alias domains
	if c.Domains != nil {
		for _, alias := range c.Domains.Aliases {
			if alias.SSL {
				return true
			}
		}
	}

	return false
}

// AddAlias adds a new alias domain to the project
func (c *Config) AddAlias(domain string, ssl bool) error {
	// Validate domain format (basic check)
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Initialize domains if nil
	if c.Domains == nil {
		c.Domains = &Domains{}
	}

	// Check if alias already exists
	for _, alias := range c.Domains.Aliases {
		if alias.Domain == domain {
			return fmt.Errorf("alias domain %s already exists", domain)
		}
	}

	// Add the new alias
	c.Domains.Aliases = append(c.Domains.Aliases, DomainAlias{
		Domain: domain,
		SSL:    ssl,
	})

	return nil
}

// RemoveAlias removes an alias domain from the project
func (c *Config) RemoveAlias(domain string) error {
	if c.Domains == nil {
		return fmt.Errorf("no alias domains configured")
	}

	// Find and remove the alias
	for i, alias := range c.Domains.Aliases {
		if alias.Domain == domain {
			c.Domains.Aliases = append(c.Domains.Aliases[:i], c.Domains.Aliases[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("alias domain %s not found", domain)
}

// FindByDomain locates a project configuration by any domain (primary or alias)
func FindByDomain(baseDir, domain string) (Config, Layout, error) {
	if baseDir == "" {
		return Config{}, Layout{}, fmt.Errorf("projects base directory is empty")
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, Layout{}, fmt.Errorf("%w: domain %s", ErrProjectNotFound, domain)
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

		// Check primary domain
		if cfg.Site != nil && cfg.Site.Domain == domain {
			if cfg.Runtime.PHPFPM == "" {
				cfg.Runtime.PHPFPM = layout.SocketPath
			}
			return cfg, layout, nil
		}

		// Check alias domains
		if cfg.Domains != nil {
			for _, alias := range cfg.Domains.Aliases {
				if alias.Domain == domain {
					if cfg.Runtime.PHPFPM == "" {
						cfg.Runtime.PHPFPM = layout.SocketPath
					}
					return cfg, layout, nil
				}
			}
		}
	}

	return Config{}, Layout{}, fmt.Errorf("%w: domain %s", ErrProjectNotFound, domain)
}
