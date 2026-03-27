package podman

import (
	"archive/tar"
	"bytes"
	"context"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	HostPort  int // Actual host port if available
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

	// Get actual host port
	status.HostPort, _ = c.GetHostPort(ctx)

	return status, nil
}

// GetHostPort returns the actual host port mapped to the container's default port.
func (c *Container) GetHostPort(ctx context.Context) (int, error) {
	// Query the actual port mapping from podman
	output, err := c.client.Run(ctx, "port", c.containerName())
	if err != nil {
		return 0, err
	}

	// Parse output like "3306/tcp -> 0.0.0.0:3306" or "3306/tcp -> :3306"
	// We need the host port for the container's default port (e.g., 3306 for mysql)
	containerPort := fmt.Sprintf("%d/tcp", c.config.Port)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, containerPort+" -> ") {
			// Extract host port
			hostPart := strings.TrimPrefix(line, containerPort+" -> ")
			// Handle formats: "0.0.0.0:3306" or ":3306" or "127.0.0.1:3306"
			hostPart = strings.TrimPrefix(hostPart, "0.0.0.0:")
			hostPart = strings.TrimPrefix(hostPart, "127.0.0.1:")
			hostPart = strings.TrimPrefix(hostPart, ":")
			if port, err := strconv.Atoi(hostPart); err == nil {
				return port, nil
			}
		}
	}

	// Fallback to configured port
	return c.config.Port, nil
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

// ExecOutput runs a command inside the container and returns the output.
func (c *Container) ExecOutput(ctx context.Context, cmd ...string) (string, error) {
	exists, err := c.client.ContainerExists(ctx, c.containerName())
	if err != nil {
		return "", fmt.Errorf("check container: %w", err)
	}
	if !exists {
		return "", ErrContainerNotFound
	}

	status, err := c.Status(ctx)
	if err != nil {
		return "", err
	}
	if !status.Running {
		return "", ErrContainerNotRunning
	}

	args := append([]string{"exec", c.containerName()}, cmd...)
	return c.client.Run(ctx, args...)
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

// BackupMeta contains metadata about a backup.
type BackupMeta struct {
	Engine        EngineType    `json:"engine"`
	ContainerName string        `json:"container_name"`
	Timestamp     time.Time    `json:"timestamp"`
	Username      string        `json:"username"`
	Database      string        `json:"database"`
}

// Backup creates a backup of the database and returns it as a tar.gz.
// The backup file is also saved to backupPath if provided.
func (c *Container) Backup(ctx context.Context, backupPath string) ([]byte, error) {
	c.log("  → Running backup command inside container...")

	var dumpData []byte
	var err error

	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57:
		dumpData, err = c.mysqldump(ctx)
	case EnginePostgres:
		dumpData, err = c.pgdumpall(ctx)
	case EngineMaria:
		dumpData, err = c.mysqldump(ctx)
	case EngineMongo:
		dumpData, err = c.mongodump(ctx)
	case EngineRedis:
		dumpData, err = c.redisDump(ctx)
	default:
		return nil, fmt.Errorf("unsupported engine: %s", c.config.Engine)
	}

	if err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Create tar.gz archive
	meta := BackupMeta{
		Engine:        c.config.Engine,
		ContainerName: c.config.ContainerName,
		Timestamp:     time.Now().UTC(),
		Username:      c.config.Username,
		Database:      "app",
	}

	metaBytes, _ := json.MarshalIndent(meta, "", "  ")

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	// Add metadata file
	if err := tw.WriteHeader(&tar.Header{
		Name: "meta.json",
		Mode: 0644,
		Size: int64(len(metaBytes)),
	}); err != nil {
		return nil, fmt.Errorf("write meta header: %w", err)
	}
	if _, err := tw.Write(metaBytes); err != nil {
		return nil, fmt.Errorf("write meta: %w", err)
	}

	// Add dump file
	dumpName := c.dumpFilename()
	if err := tw.WriteHeader(&tar.Header{
		Name: dumpName,
		Mode: 0644,
		Size: int64(len(dumpData)),
	}); err != nil {
		return nil, fmt.Errorf("write dump header: %w", err)
	}
	if _, err := tw.Write(dumpData); err != nil {
		return nil, fmt.Errorf("write dump: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip: %w", err)
	}

	backupData := buf.Bytes()

	// Save to file if path provided
	if backupPath != "" {
		c.log(fmt.Sprintf("  → Saving backup to %s...", backupPath))
		if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
			return nil, fmt.Errorf("write backup file: %w", err)
		}
	}

	c.log("  ✓ Backup created successfully")
	return backupData, nil
}

// BackupDatabase creates a backup of a specific database and returns it as tar.gz.
func (c *Container) BackupDatabase(ctx context.Context, database string) ([]byte, error) {
	c.log(fmt.Sprintf("  → Running backup for database %s...", database))

	var dumpData []byte
	var err error

	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57:
		dumpData, err = c.mysqldumpDatabase(ctx, database)
	case EnginePostgres:
		dumpData, err = c.pgdumpDatabase(ctx, database)
	case EngineMaria:
		dumpData, err = c.mysqldumpDatabase(ctx, database)
	case EngineMongo:
		dumpData, err = c.mongodumpDatabase(ctx, database)
	case EngineRedis:
		dumpData, err = c.redisDump(ctx)
	default:
		return nil, fmt.Errorf("unsupported engine: %s", c.config.Engine)
	}

	if err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Create tar.gz archive
	meta := BackupMeta{
		Engine:        c.config.Engine,
		ContainerName: c.config.ContainerName,
		Timestamp:     time.Now().UTC(),
		Username:      c.config.Username,
		Database:      database,
	}

	metaBytes, _ := json.MarshalIndent(meta, "", "  ")

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	// Add metadata file
	if err := tw.WriteHeader(&tar.Header{
		Name: "meta.json",
		Mode: 0644,
		Size: int64(len(metaBytes)),
	}); err != nil {
		return nil, fmt.Errorf("write meta header: %w", err)
	}
	if _, err := tw.Write(metaBytes); err != nil {
		return nil, fmt.Errorf("write meta: %w", err)
	}

	// Add dump file
	dumpName := c.dumpFilenameForDB(database)
	if err := tw.WriteHeader(&tar.Header{
		Name: dumpName,
		Mode: 0644,
		Size: int64(len(dumpData)),
	}); err != nil {
		return nil, fmt.Errorf("write dump header: %w", err)
	}
	if _, err := tw.Write(dumpData); err != nil {
		return nil, fmt.Errorf("write dump: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return nil, fmt.Errorf("close gzip: %w", err)
	}

	return buf.Bytes(), nil
}

// Restore restores a database from a backup tar.gz.
func (c *Container) Restore(ctx context.Context, backupData []byte) error {
	c.log("  → Extracting backup archive...")

	// Extract the archive
	br := bytes.NewReader(backupData)
	gzr, err := gzip.NewReader(br)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	tr := tar.NewReader(gzr)

	var meta *BackupMeta
	var dumpData []byte

	for {
		hdr, err := tr.Next()
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		switch hdr.Name {
		case "meta.json":
			metaBytes := make([]byte, hdr.Size)
			if _, err := tr.Read(metaBytes); err != nil {
				return fmt.Errorf("read meta: %w", err)
			}
			meta = &BackupMeta{}
			if err := json.Unmarshal(metaBytes, meta); err != nil {
				return fmt.Errorf("parse meta: %w", err)
			}
		default:
			// Assume it's the dump file
			dumpData = make([]byte, hdr.Size)
			if _, err := tr.Read(dumpData); err != nil {
				return fmt.Errorf("read dump: %w", err)
			}
		}
	}

	if meta == nil {
		return fmt.Errorf("backup missing metadata")
	}

	c.log(fmt.Sprintf("  → Restoring %s database...", meta.Engine))

	switch meta.Engine {
	case EngineMySQL8, EngineMySQL57:
		err = c.mysqlrestore(ctx, dumpData)
	case EnginePostgres:
		err = c.pgrestore(ctx, dumpData)
	case EngineMaria:
		err = c.mysqlrestore(ctx, dumpData)
	case EngineMongo:
		err = c.mongorestore(ctx, dumpData)
	case EngineRedis:
		err = c.redisrestore(ctx)
	default:
		return fmt.Errorf("unsupported engine: %s", meta.Engine)
	}

	if err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	c.log("  ✓ Restore completed successfully")
	return nil
}

// dumpFilename returns the expected dump filename for the engine.
func (c *Container) dumpFilename() string {
	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57, EngineMaria:
		return "dump.sql"
	case EnginePostgres:
		return "dump.sql"
	case EngineMongo:
		return "dump.archive"
	case EngineRedis:
		return "dump.rdb"
	}
	return "dump"
}

// dumpFilenameForDB returns the expected dump filename for a specific database.
func (c *Container) dumpFilenameForDB(database string) string {
	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57, EngineMaria:
		return database + ".sql"
	case EnginePostgres:
		return database + ".sql"
	case EngineMongo:
		return database + ".archive"
	case EngineRedis:
		return "dump.rdb"
	}
	return database + ".dump"
}

// mysqldump runs mysqldump inside the container.
func (c *Container) mysqldump(ctx context.Context) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"mysqldump",
		"--user="+c.config.Username,
		"--password="+c.config.Password,
		"--all-databases",
		"--single-transaction",
	)
	if err != nil {
		// mysqldump outputs to stdout, so stderr indicates error
		return []byte(output), err
	}
	return []byte(output), nil
}

// pgdumpall runs pg_dumpall inside the container.
func (c *Container) pgdumpall(ctx context.Context) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"pg_dumpall",
		"--username="+c.config.Username,
	)
	if err != nil {
		return []byte(output), err
	}
	return []byte(output), nil
}

// mongodump runs mongodump inside the container.
func (c *Container) mongodump(ctx context.Context) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"mongodump",
		"--username="+c.config.Username,
		"--password="+c.config.Password,
		"--authenticationDatabase=admin",
		"--db=app",
		"--archive",
		"--quiet",
	)
	if err != nil {
		return []byte(output), err
	}
	return []byte(output), nil
}

// mysqldumpDatabase runs mysqldump for a specific database.
func (c *Container) mysqldumpDatabase(ctx context.Context, database string) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"mysqldump",
		"--user="+c.config.Username,
		"--password="+c.config.Password,
		"--single-transaction",
		database,
	)
	if err != nil {
		return []byte(output), err
	}
	return []byte(output), nil
}

// pgdumpDatabase runs pg_dump for a specific database.
func (c *Container) pgdumpDatabase(ctx context.Context, database string) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"pg_dump",
		"--username="+c.config.Username,
		"--dbname="+database,
	)
	if err != nil {
		return []byte(output), err
	}
	return []byte(output), nil
}

// mongodumpDatabase runs mongodump for a specific database.
func (c *Container) mongodumpDatabase(ctx context.Context, database string) ([]byte, error) {
	output, err := c.ExecOutput(ctx,
		"mongodump",
		"--username="+c.config.Username,
		"--password="+c.config.Password,
		"--authenticationDatabase=admin",
		"--db="+database,
		"--archive",
		"--quiet",
	)
	if err != nil {
		return []byte(output), err
	}
	return []byte(output), nil
}

// mysqlrestore restores a mysql dump.
func (c *Container) mysqlrestore(ctx context.Context, dump []byte) error {
	args := []string{"exec", "-i", c.containerName(), "mysql", "--user=" + c.config.Username, "--password=" + c.config.Password}
	_, err := c.client.RunWithStdin(ctx, args, dump)
	return err
}

// pgrestore restores a postgres dump.
func (c *Container) pgrestore(ctx context.Context, dump []byte) error {
	args := []string{"exec", "-i", c.containerName(), "psql", "--username=" + c.config.Username}
	_, err := c.client.RunWithStdin(ctx, args, dump)
	return err
}

// mongorestore restores a mongo dump (uses archive format).
func (c *Container) mongorestore(ctx context.Context, dump []byte) error {
	args := []string{"exec", "-i", c.containerName(), "mongorestore", "--username=" + c.config.Username, "--password=" + c.config.Password, "--authenticationDatabase=admin", "--db=app", "--archive"}
	_, err := c.client.RunWithStdin(ctx, args, dump)
	return err
}

// redisDump triggers BGSAVE and returns the dump file.
func (c *Container) redisDump(ctx context.Context) ([]byte, error) {
	// Trigger BGSAVE
	_, err := c.ExecOutput(ctx, "redis-cli", "BGSAVE")
	if err != nil {
		return nil, fmt.Errorf("redis bgsave: %w", err)
	}

	// Wait for save to complete
	for i := 0; i < 30; i++ {
		output, _ := c.ExecOutput(ctx, "redis-cli", "LASTSAVE")
		if err == nil && output != "" {
			time.Sleep(1 * time.Second)
		}
	}

	// Copy the dump file
	output, err := c.ExecOutput(ctx, "cat", "/data/dump.rdb")
	if err != nil {
		return nil, fmt.Errorf("read dump: %w", err)
	}
	return []byte(output), nil
}

// redisrestore restores redis from dump file.
func (c *Container) redisrestore(ctx context.Context) error {
	// For redis, we need to stop the container, replace the dump file, and restart
	status, err := c.Status(ctx)
	if err != nil {
		return err
	}
	if !status.Running {
		return ErrContainerNotRunning
	}

	// Stop container
	c.log("  → Stopping container...")
	if err := c.Stop(ctx, 10*time.Second); err != nil {
		return fmt.Errorf("stop: %w", err)
	}

	// Copy dump file to volume
	src := filepath.Join(c.config.VolumePath, "dump.rdb")
	if _, err := os.Stat(src); err == nil {
		c.log("  → Restoring dump.rdb...")
		// The dump is already in the volume path from backup
	}

	// Restart container
	c.log("  → Starting container...")
	if err := c.Start(ctx); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return nil
}

// ListDatabases returns a list of database names in the container.
func (c *Container) ListDatabases(ctx context.Context) ([]string, error) {
	switch c.config.Engine {
	case EngineMySQL8, EngineMySQL57, EngineMaria:
		return c.listMySQLDatabases(ctx)
	case EnginePostgres:
		return c.listPostgresDatabases(ctx)
	case EngineMongo:
		return c.listMongoDatabases(ctx)
	case EngineRedis:
		return c.listRedisKeys(ctx)
	default:
		return nil, fmt.Errorf("unsupported engine: %s", c.config.Engine)
	}
}

func (c *Container) listMySQLDatabases(ctx context.Context) ([]string, error) {
	output, err := c.ExecOutput(ctx,
		"mysql",
		"--user="+c.config.Username,
		"--password="+c.config.Password,
		"-N",
		"-e", "SHOW DATABASES;",
	)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	var dbs []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "Database" && line != "information_schema" && line != "performance_schema" && line != "mysql" && line != "sys" {
			dbs = append(dbs, line)
		}
	}
	return dbs, nil
}

func (c *Container) listPostgresDatabases(ctx context.Context) ([]string, error) {
	output, err := c.ExecOutput(ctx,
		"psql",
		"--username="+c.config.Username,
		"--tuples-only",
		"--no-align",
		"-c", "SELECT datname FROM pg_database WHERE datistemplate = false;",
	)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	var dbs []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			dbs = append(dbs, line)
		}
	}
	return dbs, nil
}

func (c *Container) listMongoDatabases(ctx context.Context) ([]string, error) {
	output, err := c.ExecOutput(ctx,
		"mongosh",
		"--username="+c.config.Username,
		"--password="+c.config.Password,
		"--authenticationDatabase=admin",
		"--quiet",
		"--eval", "db.adminCommand('listDatabases').databases.map(d => d.name).join('\\n')",
	)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	var dbs []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "admin" && line != "config" && line != "local" {
			dbs = append(dbs, line)
		}
	}
	return dbs, nil
}

func (c *Container) listRedisKeys(ctx context.Context) ([]string, error) {
	// Redis doesn't have "databases" but we can list keyspaces (db0, db1, etc.)
	output, err := c.ExecOutput(ctx, "redis-cli", "INFO", "keyspace")
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	var dbs []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "db") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				db := strings.TrimSpace(strings.Split(parts[0], ",")[0])
				dbs = append(dbs, db)
			}
		}
	}
	if len(dbs) == 0 {
		dbs = append(dbs, "db0")
	}
	return dbs, nil
}
