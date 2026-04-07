package panel

import (
	"encoding/json"
	"net/http"
)

type ContainerResponse struct {
	Name      string `json:"name"`
	Engine    string `json:"engine"`
	Status    string `json:"status"`
	HostPort  int    `json:"hostPort"`
	CreatedAt string `json:"createdAt"`
}

type ContainerDetailResponse struct {
	ContainerResponse
	Config ContainerConfigResponse `json:"config"`
}

type ContainerConfigResponse struct {
	DatabaseUser     string `json:"databaseUser"`
	DatabaseName     string `json:"databaseName"`
	DatabasePassword string `json:"databasePassword"`
}

type BackupResponse struct {
	Name      string `json:"name"`
	Container string `json:"container"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
}

type DatabaseBackup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BackupCreateRequest struct {
	Container string           `json:"container"`
	Databases []DatabaseBackup `json:"databases"`
}

type BackupCreateResponse struct {
	Message string   `json:"message"`
	Backups []string `json:"backups"`
}

type DatabasesResponse struct {
	Databases []string `json:"databases"`
}

type BackupRestoreRequest struct {
	Container string `json:"container"`
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
