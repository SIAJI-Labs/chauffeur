package runtime

import "testing"

func TestContainerDatabaseHost(t *testing.T) {
	for _, test := range []struct {
		input, want string
	}{
		{"localhost", "host.containers.internal"},
		{"  '127.0.0.1'  ", "host.containers.internal"},
		{"db.internal", "db.internal"},
		{"", ""},
	} {
		if got := ContainerDatabaseHost(test.input); got != test.want {
			t.Errorf("ContainerDatabaseHost(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestContainerDatabaseEnvIncludesPortOnlyForLoopback(t *testing.T) {
	got := ContainerDatabaseEnv([]byte("DB_HOST=localhost\nDB_PORT=16379\n"))
	if len(got) != 2 || got[0] != "DB_HOST=host.containers.internal" || got[1] != "DB_PORT=16379" {
		t.Fatalf("ContainerDatabaseEnv() = %#v", got)
	}
	if got := ContainerDatabaseEnv([]byte("DB_HOST=db.internal\nDB_PORT=5432\n")); got != nil {
		t.Fatalf("non-loopback env overrides = %#v, want nil", got)
	}
}
