package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/siegg/chauffeur/internal/podman"
)

type Server struct {
	addr           string
	port           int
	httpServer     *http.Server
	client         *podman.PodmanClient
	devFrontendURL string
}

func logErrorValue(err error) string {
	if err == nil {
		return "<nil>"
	}

	return err.Error()
}

func logListDatabasesResult(containerName, engine string, databases []string, err error) {
	if err != nil {
		log.Printf("panel: list databases container=%q engine=%q count=%d error=%q", containerName, engine, len(databases), logErrorValue(err))
		return
	}

	log.Printf("panel: list databases container=%q engine=%q count=%d", containerName, engine, len(databases))
}

func logCreateBackupRequest(containerName string, databases []DatabaseBackup) {
	names := make([]string, 0, len(databases))
	for _, db := range databases {
		names = append(names, db.Name)
	}

	log.Printf("panel: create backup container=%q requested_databases=%v", containerName, names)
}

func logCreateBackupStart(containerName, database string) {
	log.Printf("panel: backup start container=%q database=%q", containerName, database)
}

func logCreateBackupSuccess(containerName, database, filename string) {
	log.Printf("panel: backup success container=%q database=%q filename=%q", containerName, database, filename)
}

func logCreateBackupFailure(containerName, database string, err error) {
	log.Printf("panel: backup failure container=%q database=%q error=%q", containerName, database, logErrorValue(err))
}

func logCreateBackupSummary(containerName string, succeeded, failed int) {
	log.Printf("panel: create backup summary container=%q succeeded=%d failed=%d", containerName, succeeded, failed)
}

func NewServer(port int) *Server {
	return NewServerWithAddr(fmt.Sprintf("127.0.0.1:%d", port))
}

func NewServerWithAddr(addr string) *Server {
	return &Server{
		addr:   addr,
		client: podman.NewPodmanClient(),
	}
}

func NewDevServerWithAddr(addr, frontendURL string) *Server {
	server := NewServerWithAddr(addr)
	server.devFrontendURL = strings.TrimRight(frontendURL, "/")
	return server
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/containers", s.handleListContainers)
	mux.HandleFunc("GET /api/containers/{name}", s.handleGetContainer)
	mux.HandleFunc("POST /api/containers/{name}/start", s.handleStartContainer)
	mux.HandleFunc("POST /api/containers/{name}/stop", s.handleStopContainer)
	mux.HandleFunc("GET /api/containers/{name}/logs", s.handleContainerLogs)
	mux.HandleFunc("GET /api/containers/{name}/databases", s.handleListDatabases)
	mux.HandleFunc("GET /api/backups", s.handleListBackups)
	mux.HandleFunc("POST /api/backups", s.handleCreateBackup)
	mux.HandleFunc("DELETE /api/backups/{name}", s.handleDeleteBackup)
	mux.HandleFunc("POST /api/backups/{name}/restore", s.handleRestoreBackup)

	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /assets/", s.handleAssets)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Panel server starting on http://%s", s.addr)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-serveErr:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("listen on %s: %w", s.addr, err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"version":   "0.1.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleListContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.client.Available(ctx); err != nil {
		if err == podman.ErrPodmanNotFound {
			jsonError(w, http.StatusServiceUnavailable, "Podman is not available")
			return
		}
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	engines, err := podman.ListEngines()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	containers := make([]ContainerResponse, 0, len(engines))
	for _, e := range engines {
		cfg, err := podman.Load(e)
		if err != nil || cfg == nil {
			continue
		}

		container := podman.NewContainer(s.client, cfg)
		running, _ := container.IsRunning(ctx)

		status := "stopped"
		if running {
			status = "running"
		}

		hostPort := 0
		if cfg.Port > 0 {
			hostPort = cfg.Port
		}

		containers = append(containers, ContainerResponse{
			Name:      e,
			Engine:    string(cfg.Engine),
			Status:    status,
			HostPort:  hostPort,
			CreatedAt: cfg.CreatedAt,
		})
	}

	jsonResponse(w, http.StatusOK, containers)
}

func (s *Server) handleGetContainer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	cfg, err := podman.Load(name)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)
	running, _ := container.IsRunning(ctx)

	status := "stopped"
	if running {
		status = "running"
	}

	hostPort := 0
	if cfg.Port > 0 {
		hostPort = cfg.Port
	}

	resp := ContainerDetailResponse{
		ContainerResponse: ContainerResponse{
			Name:      name,
			Engine:    string(cfg.Engine),
			Status:    status,
			HostPort:  hostPort,
			CreatedAt: cfg.CreatedAt,
		},
		Config: ContainerConfigResponse{
			DatabaseUser:     cfg.Username,
			DatabaseName:     "app",
			DatabasePassword: cfg.Password,
		},
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) handleStartContainer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	cfg, err := podman.Load(name)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)
	running, _ := container.IsRunning(ctx)
	if running {
		jsonError(w, http.StatusBadRequest, "Container already running")
		return
	}

	if err := container.Start(ctx); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Container started"})
}

func (s *Server) handleStopContainer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	cfg, err := podman.Load(name)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)
	running, _ := container.IsRunning(ctx)
	if !running {
		jsonError(w, http.StatusBadRequest, "Container not running")
		return
	}

	if err := container.Stop(ctx, 10*time.Second); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Container stopped"})
}

func (s *Server) handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	cfg, err := podman.Load(name)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)

	ctx := r.Context()
	logs, err := container.Logs(ctx, 100)
	if err != nil {
		logs = "Error fetching logs: " + err.Error()
	}

	setSSEHeaders(w)
	fmt.Fprintf(w, "data: %s\n\n", escapeSSE(logs))
	w.(http.Flusher).Flush()
}

func (s *Server) handleListDatabases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	cfg, err := podman.Load(name)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)
	running, _ := container.IsRunning(ctx)
	if !running {
		jsonError(w, http.StatusBadRequest, "Container not running")
		return
	}

	databases, err := container.ListDatabases(ctx)
	logListDatabasesResult(cfg.ContainerName, string(cfg.Engine), databases, err)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, DatabasesResponse{Databases: databases})
}

func (s *Server) handleListBackups(w http.ResponseWriter, r *http.Request) {
	backups, err := listBackups()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, backups)
}

func (s *Server) handleCreateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req BackupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Databases) == 0 {
		jsonError(w, http.StatusBadRequest, "No databases selected")
		return
	}

	cfg, err := podman.Load(req.Container)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)
	running, _ := container.IsRunning(ctx)
	if !running {
		jsonError(w, http.StatusBadRequest, "Container not running")
		return
	}

	logCreateBackupRequest(cfg.ContainerName, req.Databases)

	timestamp := time.Now().Format("20060102-150405")
	backupDir := backupPath(cfg.ContainerName, "")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		jsonError(w, http.StatusInternalServerError, "Could not create backup directory")
		return
	}

	var createdBackups []string
	failedBackups := 0
	for _, db := range req.Databases {
		logCreateBackupStart(cfg.ContainerName, db.Name)
		backupData, err := container.BackupDatabaseWithDescription(ctx, db.Name, db.Description)
		if err != nil {
			failedBackups++
			logCreateBackupFailure(cfg.ContainerName, db.Name, err)
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s.tar.gz", cfg.ContainerName, db.Name, timestamp)
		outputPath := filepath.Join(backupDir, filename)
		if err := os.WriteFile(outputPath, backupData, 0644); err != nil {
			failedBackups++
			logCreateBackupFailure(cfg.ContainerName, db.Name, err)
			continue
		}
		logCreateBackupSuccess(cfg.ContainerName, db.Name, filename)
		createdBackups = append(createdBackups, filename)
	}

	logCreateBackupSummary(cfg.ContainerName, len(createdBackups), failedBackups)

	if len(createdBackups) == 0 {
		jsonError(w, http.StatusInternalServerError, "All backups failed")
		return
	}

	jsonResponse(w, http.StatusOK, BackupCreateResponse{
		Message: fmt.Sprintf("Backed up %d database(s)", len(createdBackups)),
		Backups: createdBackups,
	})
}

func (s *Server) handleDeleteBackup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	backupPath := backupPath("", name)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		jsonError(w, http.StatusNotFound, "Backup not found")
		return
	}

	if err := os.Remove(backupPath); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Backup deleted"})
}

func (s *Server) handleRestoreBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	var req BackupRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	cfg, err := podman.Load(req.Container)
	if err != nil || cfg == nil {
		jsonError(w, http.StatusNotFound, "Container not found")
		return
	}

	container := podman.NewContainer(s.client, cfg)

	backupPath := backupPath(req.Container, name)
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		jsonError(w, http.StatusNotFound, "Backup not found")
		return
	}

	if err := container.Restore(ctx, backupData); err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Restore completed"})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if s.devFrontendURL != "" {
		http.Redirect(w, r, s.devFrontendURL+r.URL.RequestURI(), http.StatusTemporaryRedirect)
		return
	}
	html, err := indexHTML()
	if err != nil {
		http.Error(w, "Failed to load index", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	data, err := staticFS.Open(path[1:])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer data.Close()

	content, err := io.ReadAll(data)
	if err != nil {
		http.Error(w, "Internal error", 500)
		return
	}

	contentType := "text/plain"
	switch filepath.Ext(path) {
	case ".html":
		contentType = "text/html"
	case ".js", ".mjs":
		contentType = "application/javascript"
	case ".css":
		contentType = "text/css"
	case ".json":
		contentType = "application/json"
	case ".svg":
		contentType = "image/svg+xml"
	case ".png":
		contentType = "image/png"
	case ".ico":
		contentType = "image/x-icon"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(content)
}

func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func escapeSSE(data string) string {
	data = strings.ReplaceAll(data, "\n", "\\n")
	data = strings.ReplaceAll(data, "\r", "\\r")
	return data
}

func listBackups() ([]BackupResponse, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	backupDir := filepath.Join(home, ".chauffeur", "podman", "backups")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupResponse{}, nil
		}
		return nil, err
	}

	var backups []BackupResponse
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupFilePath := backupPath("", entry.Name())
		meta, err := podman.ReadBackupMeta(backupFilePath)
		container := entry.Name()
		if err == nil && meta.ContainerName != "" {
			container = meta.ContainerName
		}

		backups = append(backups, BackupResponse{
			Name:      entry.Name(),
			Container: container,
			Size:      info.Size(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
		})
	}

	return backups, nil
}

func backupPath(container, filename string) string {
	home, _ := os.UserHomeDir()
	if filename == "" {
		return filepath.Join(home, ".chauffeur", "podman", "backups")
	}
	return filepath.Join(home, ".chauffeur", "podman", "backups", filename)
}
