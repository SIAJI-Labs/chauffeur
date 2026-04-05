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
	addr       string
	port       int
	httpServer *http.Server
	client     *podman.PodmanClient
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

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/containers", s.handleListContainers)
	mux.HandleFunc("GET /api/containers/{name}", s.handleGetContainer)
	mux.HandleFunc("POST /api/containers/{name}/start", s.handleStartContainer)
	mux.HandleFunc("POST /api/containers/{name}/stop", s.handleStopContainer)
	mux.HandleFunc("GET /api/containers/{name}/logs", s.handleContainerLogs)
	mux.HandleFunc("GET /api/backups", s.handleListBackups)
	mux.HandleFunc("POST /api/backups", s.handleCreateBackup)
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

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	return s.httpServer.Shutdown(context.Background())
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

	backupPath := backupPath(cfg.ContainerName, "")
	_, err = container.BackupDatabaseWithDescription(ctx, "app", req.Description)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{
		"message":    "Backup created",
		"backupPath": backupPath,
	})
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

		parts := strings.Split(entry.Name(), "-")
		container := parts[0]

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
