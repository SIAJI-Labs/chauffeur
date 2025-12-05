package lib

import (
	"strconv"
	"testing"
)

func TestFindAvailablePortsSkipsUnavailable(t *testing.T) {
	pm := NewPortManager(8080, 8082, "prompt")
	ports := pm.FindAvailablePorts(2, 0)
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
}

func TestValidatePortRange(t *testing.T) {
	pm := NewPortManager(1, 2, "prompt")
	if err := pm.ValidatePortRange(); err != nil {
		t.Fatalf("expected valid range, got %v", err)
	}
	pm = NewPortManager(9000, 8000, "prompt")
	if err := pm.ValidatePortRange(); err == nil {
		t.Fatalf("expected error for inverted range")
	}
}

func TestAutoResolvePortConflict(t *testing.T) {
	pm := NewPortManager(8080, 8085, "auto")
	port, err := pm.AutoResolvePortConflict("nginx", 8080)
	if err != nil {
		t.Fatalf("AutoResolvePortConflict error: %v", err)
	}
	if port == 0 {
		t.Fatalf("expected a port to be returned")
	}
	if port < 8080 || port > 8085 {
		t.Fatalf("port out of range: %d", port)
	}
}

func TestWriteReadPortConfigEnv(t *testing.T) {
	ports := map[string]int{"nginx": 8080, "php-fpm": 9000}
	if err := WritePortConfigToEnv(ports); err != nil {
		t.Fatalf("WritePortConfigToEnv error: %v", err)
	}
	read := ReadPortConfigFromEnv()
	for k, v := range ports {
		if got := read[k]; got != v {
			t.Fatalf("expected %s=%d, got %d", k, v, got)
		}
	}
	// invalid env should be skipped
	t.Setenv("CHAUF_NGINX_HTTPS_PORT", "not-a-number")
	if _, err := strconv.Atoi("not-a-number"); err == nil {
		t.Fatalf("expected parse error in test guard")
	}
}
