package podman

import (
	"strings"
	"testing"
)

func TestRedisDefaultsDisableAuthentication(t *testing.T) {
	cfg := DefaultConfig(EngineRedis)

	if cfg.AuthEnabled {
		t.Fatal("Redis authentication should be disabled by default")
	}
	if cfg.Username != "" || cfg.Password != "" {
		t.Fatalf("Redis credentials = %q/%q, want empty", cfg.Username, cfg.Password)
	}
}

func TestRedisConfigRoundTripsAuthentication(t *testing.T) {
	cfg := &DatabaseConfig{
		Name:          "redis",
		Engine:        EngineRedis,
		Image:         "docker.io/library/redis:7-alpine",
		ContainerName: "chauf-redis",
		AuthEnabled:   true,
		Username:      "app",
		Password:      "secret",
		Port:          6379,
		VolumePath:    "/tmp/redis",
	}

	encoded := marshalConfig(cfg)
	decoded, err := unmarshalConfig(encoded)
	if err != nil {
		t.Fatalf("unmarshalConfig() error = %v", err)
	}
	if !decoded.AuthEnabled || decoded.Username != cfg.Username || decoded.Password != cfg.Password {
		t.Fatalf("decoded auth config = %#v, want enabled credentials", decoded)
	}
}

func TestLegacyRedisConfigRemainsUnauthenticated(t *testing.T) {
	decoded, err := unmarshalConfig(strings.Join([]string{
		"name: redis",
		"engine: redis",
		"username: chauf",
		"password: legacy-secret",
	}, "\n"))
	if err != nil {
		t.Fatalf("unmarshalConfig() error = %v", err)
	}
	if decoded.AuthEnabled {
		t.Fatal("legacy Redis config should not enable authentication implicitly")
	}
}

func TestRedisDSNIncludesCredentialsOnlyWhenEnabled(t *testing.T) {
	cfg := DefaultConfig(EngineRedis)
	if got, want := DSN(cfg), "redis://localhost:6379"; got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}

	cfg.AuthEnabled = true
	cfg.Username = "app user"
	cfg.Password = "secret/pass"
	if got, want := DSN(cfg), "redis://app+user:secret%2Fpass@localhost:6379"; got != want {
		t.Fatalf("authenticated DSN() = %q, want %q", got, want)
	}
}
