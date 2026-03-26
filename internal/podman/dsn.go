package podman

import (
	"fmt"
)

// DSN returns the connection string for the given database config.
func DSN(cfg *DatabaseConfig) string {
	switch cfg.Engine {
	case EngineMySQL8, EngineMySQL57, EngineMaria:
		return fmt.Sprintf("mysql://%s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EnginePostgres:
		return fmt.Sprintf("postgres://%s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EngineMongo:
		return fmt.Sprintf("mongodb://%s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EngineRedis:
		return fmt.Sprintf("redis://localhost:%d", cfg.Port)
	default:
		return ""
	}
}

// DSNLabel returns a human-readable label for the DSN.
func DSNLabel(cfg *DatabaseConfig) string {
	switch cfg.Engine {
	case EngineMySQL8, EngineMySQL57, EngineMaria:
		return fmt.Sprintf("MySQL %s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EnginePostgres:
		return fmt.Sprintf("PostgreSQL %s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EngineMongo:
		return fmt.Sprintf("MongoDB %s:%s@localhost:%d/app",
			cfg.Username, cfg.Password, cfg.Port)
	case EngineRedis:
		return fmt.Sprintf("Redis localhost:%d", cfg.Port)
	default:
		return ""
	}
}
