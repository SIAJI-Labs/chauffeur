package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config mirrors chauffeur.yaml.
type Config struct {
	Workspace string
	Nginx     struct {
		HTTPPort  int
		HTTPSPort int
	}
	PHP struct {
		DefaultVersion string
	}
	DNS struct {
		TLD     string
		Enabled bool
	}
	Logging struct {
		Level     string
		MaxSizeMB int
	}
	Version   string
	CreatedAt string
}

// DefaultConfig returns a Config populated with defaults for the given workspace root.
func DefaultConfig(root string) Config {
	var c Config
	c.Workspace = root
	c.Nginx.HTTPPort = 8080
	c.Nginx.HTTPSPort = 8443
	c.PHP.DefaultVersion = "8.3"
	c.DNS.TLD = "test"
	c.DNS.Enabled = true
	c.Logging.Level = "info"
	c.Logging.MaxSizeMB = 10
	c.Version = "2"
	c.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	return c
}

// Load reads chauffeur.yaml from the workspace root.
// Falls back to DefaultConfig if the file doesn't exist or can't be parsed.
func Load() Config {
	root := Root()
	c := DefaultConfig(root)

	data, err := os.ReadFile(filepath.Join(root, "config", "chauffeur.yaml"))
	if err != nil {
		return c
	}

	var section string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header: a non-indented key ending with ":" and no value
		if !strings.HasPrefix(raw, " ") && strings.HasSuffix(line, ":") && !strings.Contains(line, ": ") {
			section = strings.TrimSuffix(line, ":")
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		switch section + "." + key {
		case ".workspace":
			c.Workspace = val
		case "nginx.http_port":
			if v, err := strconv.Atoi(val); err == nil {
				c.Nginx.HTTPPort = v
			}
		case "nginx.https_port":
			if v, err := strconv.Atoi(val); err == nil {
				c.Nginx.HTTPSPort = v
			}
		case "php.default_version":
			c.PHP.DefaultVersion = val
		case "dns.tld":
			c.DNS.TLD = val
		case "dns.enabled":
			c.DNS.Enabled = val == "true"
		case "logging.level":
			c.Logging.Level = val
		case "logging.max_size_mb":
			if v, err := strconv.Atoi(val); err == nil {
				c.Logging.MaxSizeMB = v
			}
		case ".version":
			c.Version = val
		case ".created_at":
			c.CreatedAt = val
		}
	}

	return c
}

// SetDefaultPHP updates php.default_version in chauffeur.yaml.
func SetDefaultPHP(version string) error {
	root := Root()
	configPath := filepath.Join(root, "config", "chauffeur.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "default_version") {
			lines[i] = "  default_version: \"" + version + "\""
			found = true
			break
		}
	}

	if !found {
		// Insert after "php:" section header
		for i, line := range lines {
			if strings.TrimSpace(line) == "php:" {
				lines = append(lines[:i+1], append([]string{"  default_version: \"" + version + "\""}, lines[i+1:]...)...)
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("could not find php section in %s", configPath)
	}

	return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
}

// DefaultConfigYAML returns the default chauffeur.yaml content for the given workspace root.
func DefaultConfigYAML(root string) string {
	return fmt.Sprintf(`workspace: %s
nginx:
  http_port: 8080
  https_port: 8443
php:
  default_version: "8.3"
dns:
  tld: test
  enabled: true
logging:
  level: info
  max_size_mb: 10
version: "2"
created_at: "%s"
`, root, time.Now().UTC().Format(time.RFC3339))
}
