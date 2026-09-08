package podman

import (
	"reflect"
	"testing"
)

func TestRedisCLIArgsUseConfiguredAuthentication(t *testing.T) {
	c := &Container{config: &DatabaseConfig{Engine: EngineRedis}}
	if got, want := c.redisCLIArgs("PING"), []string{"redis-cli", "PING"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unauthenticated redis args = %#v, want %#v", got, want)
	}

	c.config.AuthEnabled = true
	c.config.Username = "app"
	c.config.Password = "secret"
	if got, want := c.redisCLIArgs("PING"), []string{"redis-cli", "--user", "app", "--pass", "secret", "PING"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("authenticated redis args = %#v, want %#v", got, want)
	}
}
