package panel

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed static/*
var staticFiles embed.FS

var staticFS, _ = fs.Sub(staticFiles, "static")

func indexHTML() (string, error) {
	data, err := staticFS.Open("index.html")
	if err != nil {
		return "", err
	}
	defer data.Close()
	content, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func appJS() (string, error) {
	return readAsset("assets/index-*.js")
}

func appCSS() (string, error) {
	return readAsset("assets/index-*.css")
}

func readAsset(pattern string) (string, error) {
	entries, err := fs.ReadDir(staticFS, "assets")
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		name := entry.Name()
		if matched, _ := filepath.Match(pattern, name); matched {
			data, err := staticFS.Open(filepath.Join("assets", name))
			if err != nil {
				return "", err
			}
			defer data.Close()
			content, err := io.ReadAll(data)
			if err != nil {
				return "", err
			}
			return string(content), nil
		}
	}
	return "", nil
}

func ServeStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	path = filepath.Clean(path)

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
	case ".js":
		contentType = "application/javascript"
	case ".css":
		contentType = "text/css"
	case ".json":
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}
