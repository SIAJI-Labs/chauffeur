package podman

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// EngineType represents the database engine type.
type EngineType string

const (
	EngineMySQL8   EngineType = "mysql8"
	EngineMySQL57  EngineType = "mysql57"
	EnginePostgres EngineType = "postgres"
	EngineMaria    EngineType = "maria"
	EngineMongo    EngineType = "mongo"
	EngineRedis    EngineType = "redis"
)

// IsValidEngine checks if the given string is a valid engine type.
func IsValidEngine(s string) bool {
	switch EngineType(s) {
	case EngineMySQL8, EngineMySQL57, EnginePostgres, EngineMaria, EngineMongo, EngineRedis:
		return true
	}
	return false
}

// engineDefaults holds default image and port for each engine.
var engineDefaults = map[EngineType]struct {
	Image     string
	Port      int
	VolumeDir string
}{
	EngineMySQL8: {
		Image:     "docker.io/library/mysql:8.0",
		Port:      3306,
		VolumeDir: "mysql8",
	},
	EngineMySQL57: {
		Image:     "docker.io/library/mysql:5.7",
		Port:      3307,
		VolumeDir: "mysql57",
	},
	EnginePostgres: {
		Image:     "docker.io/library/postgres:16",
		Port:      5432,
		VolumeDir: "postgres",
	},
	EngineMaria: {
		Image:     "docker.io/library/mariadb:11",
		Port:      3306,
		VolumeDir: "maria",
	},
	EngineMongo: {
		Image:     "docker.io/library/mongo:7",
		Port:      27017,
		VolumeDir: "mongo",
	},
	EngineRedis: {
		Image:     "docker.io/library/redis:7-alpine",
		Port:      6379,
		VolumeDir: "redis",
	},
}

// DatabaseConfig describes a managed database container.
type DatabaseConfig struct {
	Name          string      `json:"name"`
	Engine        EngineType  `json:"engine"`
	Image         string      `json:"image"`
	ContainerName string      `json:"container_name"`
	Username      string      `json:"username"`
	Password      string      `json:"password"`
	Port          int         `json:"port"`
	VolumePath    string      `json:"volume_path"`
	Env           []EnvVar    `json:"env"`
	CreatedAt     string      `json:"created_at"`
}

// EnvVar represents a single environment variable.
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DefaultConfig returns a DatabaseConfig with defaults for the given engine.
func DefaultConfig(engine EngineType) *DatabaseConfig {
	home, _ := os.UserHomeDir()
	defaults := engineDefaults[engine]

	username := "chauf"
	password := GeneratePassword()

	return &DatabaseConfig{
		Name:          string(engine),
		Engine:        engine,
		Image:         defaults.Image,
		ContainerName: "chauf-" + string(engine),
		Username:      username,
		Password:      password,
		Port:          defaults.Port,
		VolumePath:    filepath.Join(home, PodmanRoot, VolumesDir, "chauf-"+string(engine)),
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
}

// GeneratePassword generates a random 16-character password.
func GeneratePassword() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Load reads a database config. It first tries by container name,
// then falls back to engine name for backward compatibility.
func Load(containerOrEngine string) (*DatabaseConfig, error) {
	// Try container name first
	path := ConfigPath(containerOrEngine)
	data, err := os.ReadFile(path)
	if err == nil {
		return unmarshalConfig(string(data))
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Fall back to engine name (backward compatibility)
	path = filepath.Join(Root(), containerOrEngine+".yaml")
	data, err = os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	return unmarshalConfig(string(data))
}

// Save writes a database config using container name as the filename.
func Save(cfg *DatabaseConfig) error {
	if err := EnsureRoot(); err != nil {
		return err
	}
	if err := EnsureVolumePath(cfg.ContainerName); err != nil {
		return err
	}
	path := ConfigPath(cfg.ContainerName)
	return os.WriteFile(path, []byte(marshalConfig(cfg)), 0644)
}

// ListEngines returns all container names that have config files.
// It scans for *.yaml files and returns the filenames (without extension).
func ListEngines() ([]string, error) {
	root := Root()
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	seen := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		if name == "config" {
			continue
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names, nil
}

// ── YAML serialization ────────────────────────────────────────────────────────

func marshalConfig(cfg *DatabaseConfig) string {
	var sb strings.Builder
	line := func(s string) { sb.WriteString(s + "\n") }

	line("name: " + cfg.Name)
	line("engine: " + string(cfg.Engine))
	line("image: " + cfg.Image)
	line("container_name: " + cfg.ContainerName)
	line("username: " + cfg.Username)
	line("password: " + cfg.Password)
	line(fmt.Sprintf("port: %d", cfg.Port))
	line("volume_path: " + cfg.VolumePath)
	if len(cfg.Env) > 0 {
		line("env:")
		for _, e := range cfg.Env {
			line(fmt.Sprintf("  - key: %s", e.Key))
			line(fmt.Sprintf("    value: %s", e.Value))
		}
	}
	line("created_at: " + cfg.CreatedAt)
	return sb.String()
}

func unmarshalConfig(data string) (*DatabaseConfig, error) {
	cfg := &DatabaseConfig{}
	inEnv := false

	for _, raw := range strings.Split(data, "\n") {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		isIndented := strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t")

		// Env list items
		if inEnv && isIndented && strings.HasPrefix(trimmed, "- key:") {
			key := strings.TrimPrefix(trimmed, "- key: ")
			cfg.Env = append(cfg.Env, EnvVar{Key: key})
			continue
		}
		if inEnv && isIndented && strings.HasPrefix(trimmed, "value:") {
			val := strings.TrimPrefix(trimmed, "value: ")
			if len(cfg.Env) > 0 {
				cfg.Env[len(cfg.Env)-1].Value = val
			}
			continue
		}

		// Non-indented line resets section tracking
		if !isIndented {
			inEnv = false
		}

		// Key-value
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			cfg.Name = val
		case "engine":
			cfg.Engine = EngineType(val)
		case "image":
			cfg.Image = val
		case "container_name":
			cfg.ContainerName = val
		case "username":
			cfg.Username = val
		case "password":
			cfg.Password = val
		case "port":
			fmt.Sscanf(val, "%d", &cfg.Port)
		case "volume_path":
			cfg.VolumePath = val
		case "env":
			if val == "" {
				inEnv = true
			}
		case "created_at":
			cfg.CreatedAt = val
		}
	}

	return cfg, nil
}
