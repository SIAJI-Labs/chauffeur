package podman

import "errors"

// Podman-specific errors.
var (
	// ErrPodmanNotFound is returned when the podman binary is not found.
	ErrPodmanNotFound = errors.New("podman not found")

	// ErrContainerExists is returned when a container already exists.
	ErrContainerExists = errors.New("container already exists")

	// ErrContainerNotFound is returned when a container does not exist.
	ErrContainerNotFound = errors.New("container not found")

	// ErrNetworkNotFound is returned when the chauf-net network does not exist.
	ErrNetworkNotFound = errors.New("network chauf-net not found")

	// ErrVolumeNotFound is returned when a volume does not exist.
	ErrVolumeNotFound = errors.New("volume not found")

	// ErrContainerNotRunning is returned when trying to operate on a stopped container.
	ErrContainerNotRunning = errors.New("container not running")

	// ErrEngineNotSupported is returned when an unsupported engine is requested.
	ErrEngineNotSupported = errors.New("unsupported engine")

	// ErrConfigNotFound is returned when a database config file does not exist.
	ErrConfigNotFound = errors.New("config file not found")
)
