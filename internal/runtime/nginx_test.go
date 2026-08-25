package runtime

import (
	"context"
	"testing"
)

func TestEnsureNginxContainerValidatesPort(t *testing.T) {
	err := (Podman{Runner: &recordingRunner{}}).EnsureNginxContainer(context.Background(), "/tmp/nginx.conf", "/tmp/project", 0)
	if err == nil {
		t.Fatal("expected invalid port error")
	}
}
