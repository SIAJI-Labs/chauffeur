package podman

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Logger interface for verbose output
type Logger interface {
	Print(args ...interface{})
}

// ContainerStatus describes the current state of a container.
type ContainerStatus struct {
	Running   bool
	StartedAt time.Time
	Health    string
}

// Container wraps a podman container with its config.
type Container struct {
	client  *PodmanClient
	config  *DatabaseConfig
	network *NetworkManager
	volumes *VolumeManager
	logger  Logger
}

// NewContainer creates a new Container.
func NewContainer(client *PodmanClient, cfg *DatabaseConfig) *Container {
	return &Container{
		client:  client,
		config:  cfg,
		network: NewNetworkManager(client),
		volumes: NewVolumeManager(client),
	}
}

// SetLogger sets the logger for verbose output.
func (c *Container) SetLogger(logger Logger) {
	c.logger = logger
}

func (c *Container) log(args ...interface{}) {
	if c.logger != nil {
		c.logger.Print(args...)
	}
}

// containerName returns the container name for this container.
func (c *Container) containerName() string {
	return c.config.ContainerName
}

// volumeName returns the podman volume name for this container.
// Uses the container name so volume and container are linked 1:1.
func (c *Container) volumeName() string {
	return c.config.ContainerName
}

// image returns the container image.
func (c *Container) image() string {
	return c.config.Image
}

// env returns the environment variables as a flat slice.
func (c *Container) env() []string {
	var env []string

	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57:
		env = []string{
			"MYSQL_ROOT_PASSWORD=" + c.config.Password,
			"MYSQL_DATABASE=app",
			"MYSQL_USER=" + c.config.Username,
			"MYSQL_PASSWORD=" + c.config.Password,
		}
	case EnginePostgres:
		env = []string{
			"POSTGRES_PASSWORD=" + c.config.Password,
			"POSTGRES_USER=" + c.config.Username,
			"POSTGRES_DB=app",
		}
	case EngineMaria:
		env = []string{
			"MARIADB_ROOT_PASSWORD=" + c.config.Password,
			"MARIADB_DATABASE=app",
			"MARIADB_USER=" + c.config.Username,
			"MARIADB_PASSWORD=" + c.config.Password,
		}
	case EngineMongo:
		env = []string{
			"MONGO_INITDB_ROOT_USERNAME=" + c.config.Username,
			"MONGO_INITDB_ROOT_PASSWORD=" + c.config.Password,
		}
	case EngineRedis:
		// Redis doesn't need credentials by default
	}

	// Add any custom env vars from config
	for _, e := range c.config.Env {
		env = append(env, e.Key+"="+e.Value)
	}

	return env
}

// portMapping returns the port mapping string.
func (c *Container) portMapping() string {
	return fmt.Sprintf("%d:%d", c.config.Port, c.config.Port)
}

// Create creates and starts the database container.
func (c *Container) Create(ctx context.Context) error {
	// Step 1: Ensure network exists
	c.log("  → Ensuring network exists...")
	if err := c.network.Ensure(ctx); err != nil {
		return fmt.Errorf("ensure network: %w", err)
	}

	// Step 2: Check if container already exists
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if exists {
		return ErrContainerExists
	}

	// Step 3: Pull the image explicitly so we can report progress
	c.log(fmt.Sprintf("  → Pulling image %s...", c.image()))
	_, err = c.client.Run(ctx, "pull", c.image())
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}

	// Step 4: Check if volume already exists before we create it
	volumeExisted, err := c.volumes.Exists(ctx, c.volumeName())
	if err != nil {
		return fmt.Errorf("check volume: %w", err)
	}

	// Step 5: Ensure host directory exists for bind mount
	c.log(fmt.Sprintf("  → Creating volume directory %s...", c.config.VolumePath))
	if err := os.MkdirAll(c.config.VolumePath, 0755); err != nil {
		return fmt.Errorf("create volume dir: %w", err)
	}

	// Step 6: Ensure podman volume exists (creates if not present)
	if !volumeExisted {
		c.log(fmt.Sprintf("  → Creating podman volume %s...", c.volumeName()))
		if err := c.volumes.Ensure(ctx, c.volumeName()); err != nil {
			return fmt.Errorf("ensure volume: %w", err)
		}
	}

	// Step 7: Build and run the container
	c.log(fmt.Sprintf("  → Creating container %s...", c.containerName()))
	args := []string{
		"run",
		"--detach",
		"--name", c.containerName(),
		"--network", NetworkName,
		"--publish", c.portMapping(),
		"--volume", c.volumeName() + ":" + c.config.VolumePath,
	}

	// Add environment variables
	for _, e := range c.env() {
		args = append(args, "--env", e)
	}

	// Add image
	args = append(args, c.image())

	c.log(fmt.Sprintf("  → Starting container..."))
	_, err = c.client.Run(ctx, args...)
	if err != nil {
		// Clean up: remove podman volume if we created it, and remove the host bind-mount directory
		if !volumeExisted {
			c.volumes.Remove(ctx, c.volumeName())
			os.RemoveAll(c.config.VolumePath)
		}
		return fmt.Errorf("create container: %w", err)
	}

	c.log("  ✓ Container created successfully")
	return nil
}

// Start starts a stopped container.
func (c *Container) Start(ctx context.Context) error {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return ErrContainerNotFound
	}

	_, err = c.client.Run(ctx, "start", c.containerName())
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}
	return nil
}

// Stop stops a running container.
func (c *Container) Stop(ctx context.Context, timeout time.Duration) error {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return ErrContainerNotFound
	}

	_, err = c.client.Run(ctx, "stop", "-t", fmt.Sprintf("%d", int(timeout.Seconds())), c.containerName())
	if err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	return nil
}

// Remove removes a container.
func (c *Container) Remove(ctx context.Context, force bool) error {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return ErrContainerNotFound
	}

	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, c.containerName())

	_, err = c.client.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}

// Status returns the current status of the container.
func (c *Container) Status(ctx context.Context) (*ContainerStatus, error) {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return nil, fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return nil, ErrContainerNotFound
	}

	output, err := c.client.Run(ctx, "container", "inspect", "--format", "{{.State.Running}}|{{.State.StartedAt}}|{{.State.Healthcheck}}", c.containerName())
	if err != nil {
		return nil, fmt.Errorf("inspect container: %w", err)
	}

	parts := strings.Split(output, "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected inspect output: %s", output)
	}

	status := &ContainerStatus{
		Running: parts[0] == "true",
	}

	if len(parts) >= 2 && parts[1] != "" {
		status.StartedAt, _ = time.Parse(time.RFC3339, parts[1])
	}

	if len(parts) >= 3 && parts[2] != "" {
		status.Health = strings.TrimSpace(parts[2])
	}

	return status, nil
}

// Logs returns the container logs.
func (c *Container) Logs(ctx context.Context, lines int) (string, error) {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return "", fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return "", ErrContainerNotFound
	}

	output, err := c.client.Run(ctx, "logs", "--tail", fmt.Sprintf("%d", lines), c.containerName())
	if err != nil {
		return "", fmt.Errorf("get logs: %w", err)
	}
	return output, nil
}

// Exec runs a command inside the container.
func (c *Container) Exec(ctx context.Context, cmd ...string) error {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return ErrContainerNotFound
	}

	status, err := c.Status(ctx)
	if err != nil {
		return err
	}
	if !status.Running {
		return ErrContainerNotRunning
	}

	args := append([]string{"exec", "-it", c.containerName()}, cmd...)
	_, err = c.client.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

// IsRunning returns true if the container is running.
func (c *Container) IsRunning(ctx context.Context) (bool, error) {
	status, err := c.Status(ctx)
	if err != nil {
		if err == ErrContainerNotFound {
			return false, nil
		}
		return false, err
	}
	return status.Running, nil
}
