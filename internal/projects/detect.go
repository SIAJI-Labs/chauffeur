package projects

import (
	"os"
	"path/filepath"
	"strings"
)

const DefaultProxyPort = 3000

// DefaultProxyPortFor returns the conventional development-server port for a
// detected JavaScript toolchain. Users can always override it with --proxy-port.
func DefaultProxyPortFor(dirPath string) int {
	if fileExists(filepath.Join(dirPath, "vite.config.js")) ||
		fileExists(filepath.Join(dirPath, "vite.config.ts")) ||
		fileExists(filepath.Join(dirPath, "vite.config.mjs")) {
		return 5173
	}
	if fileExists(filepath.Join(dirPath, "angular.json")) {
		return 4200
	}
	return DefaultProxyPort
}

// Detect determines the project type by inspecting the directory contents.
func Detect(dirPath string) ProjectType {
	if fileExists(filepath.Join(dirPath, "artisan")) {
		return TypeLaravel
	}
	if fileExists(filepath.Join(dirPath, "wp-config.php")) ||
		fileExists(filepath.Join(dirPath, "wp-login.php")) {
		return TypeWordPress
	}
	if hasJavaScriptManifest(dirPath) {
		return TypeReverseProxy
	}
	if hasPHPEntryPoint(dirPath) {
		return TypePHP
	}
	return TypeUnknown
}

// DocumentRoot returns the web-accessible root directory for the project.
func DocumentRoot(dirPath string, ptype ProjectType) string {
	switch ptype {
	case TypeLaravel:
		return filepath.Join(dirPath, "public")
	case TypeWordPress:
		return dirPath
	case TypeReverseProxy:
		return dirPath
	default:
		// PHP: prefer public/ if it exists, otherwise the project root.
		pub := filepath.Join(dirPath, "public")
		if info, err := os.Stat(pub); err == nil && info.IsDir() {
			return pub
		}
		return dirPath
	}
}

func hasJavaScriptManifest(dirPath string) bool {
	markers := []string{
		"package.json",
		"vite.config.js",
		"vite.config.ts",
		"vite.config.mjs",
		"next.config.js",
		"next.config.mjs",
		"next.config.ts",
		"nuxt.config.ts",
		"astro.config.mjs",
		"angular.json",
		"svelte.config.js",
	}
	for _, marker := range markers {
		if fileExists(filepath.Join(dirPath, marker)) {
			return true
		}
	}
	return false
}

func hasPHPEntryPoint(dirPath string) bool {
	if fileExists(filepath.Join(dirPath, "index.php")) ||
		fileExists(filepath.Join(dirPath, "public", "index.php")) {
		return true
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".php") {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
