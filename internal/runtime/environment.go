package runtime

import "strings"

// ContainerDatabaseHost translates only the container-side representation of
// a host-facing database address. Project files and displayed DSNs remain
// unchanged and continue to use localhost.
func ContainerDatabaseHost(host string) string {
	host = strings.Trim(strings.TrimSpace(host), `"'`)
	if host == "localhost" || host == "127.0.0.1" {
		return "host.containers.internal"
	}
	return host
}

// ContainerDatabaseEnv returns only database variables that need to differ
// inside a container. The project environment remains host-facing.
func ContainerDatabaseEnv(data []byte) []string {
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		for _, key := range []string{"DB_HOST", "DB_PORT"} {
			if strings.HasPrefix(line, key+"=") {
				values[key] = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, key+"=")), `"'`)
			}
		}
	}
	host := ContainerDatabaseHost(values["DB_HOST"])
	if host == values["DB_HOST"] {
		return nil
	}
	overrides := []string{"DB_HOST=" + host}
	if port := values["DB_PORT"]; port != "" {
		overrides = append(overrides, "DB_PORT="+port)
	}
	return overrides
}
