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
		Versions       map[string]PHPVersionConfig
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

// PHPVersionConfig holds runtime settings for a specific PHP version.
type PHPVersionConfig struct {
	UploadMaxFilesize string
	PostMaxSize       string
	MemoryLimit       string
	MaxExecutionTime  int
	MaxInputVars      int
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
	c.PHP.Versions = make(map[string]PHPVersionConfig)
	c.Version = "2"
	c.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	return c
}

// DefaultPHPVersionConfig returns the default PHP runtime settings.
func DefaultPHPVersionConfig() PHPVersionConfig {
	return PHPVersionConfig{
		UploadMaxFilesize: "64M",
		PostMaxSize:       "64M",
		MemoryLimit:       "256M",
		MaxExecutionTime:  300,
		MaxInputVars:      5000,
	}
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

		// Handle php version-specific keys like "8.3.upload_max_filesize"
		if strings.Contains(key, ".") {
			// Format: "8.3.upload_max_filesize" -> version="8.3", setting="upload_max_filesize"
			// Need SplitN with limit 3 to get version.settings correctly
			parts := strings.SplitN(key, ".", 3)
			if len(parts) == 3 {
				version := parts[0] + "." + parts[1]
				setting := parts[2]

				// Initialize version config if not exists
				if _, ok := c.PHP.Versions[version]; !ok {
					c.PHP.Versions[version] = DefaultPHPVersionConfig()
				}

				// Set the appropriate field
				cfg := c.PHP.Versions[version]
				switch setting {
				case "upload_max_filesize":
					cfg.UploadMaxFilesize = val
				case "post_max_size":
					cfg.PostMaxSize = val
				case "memory_limit":
					cfg.MemoryLimit = val
				case "max_execution_time":
					if v, err := strconv.Atoi(val); err == nil {
						cfg.MaxExecutionTime = v
					}
				case "max_input_vars":
					if v, err := strconv.Atoi(val); err == nil {
						cfg.MaxInputVars = v
					}
				}
				c.PHP.Versions[version] = cfg
			}
			continue
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

// SavePHPVersionSetting updates a single PHP version setting in chauffeur.yaml.
// version: PHP version like "8.3"
// key: setting name like "upload_max_filesize"
// value: the value to set
func SavePHPVersionSetting(version, key, value string) error {
	root := Root()
	configPath := filepath.Join(root, "config", "chauffeur.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	keyLine := fmt.Sprintf("  %s.%s:", version, key)
	valueLine := fmt.Sprintf("  %s.%s: \"%s\"", version, key, value)

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, keyLine) {
			lines[i] = valueLine
			found = true
			break
		}
	}

	if !found {
		// Insert after "php:" section
		for i, line := range lines {
			if strings.TrimSpace(line) == "php:" {
				// Insert all lines after php: (they will be in order of insertion)
				lines = append(lines[:i+1], append([]string{valueLine}, lines[i+1:]...)...)
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
