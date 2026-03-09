package projects

import (
	"os"
	"path/filepath"
)

// Detect determines the project type by inspecting the directory contents.
func Detect(dirPath string) ProjectType {
	if fileExists(filepath.Join(dirPath, "artisan")) {
		return TypeLaravel
	}
	if fileExists(filepath.Join(dirPath, "wp-config.php")) ||
		fileExists(filepath.Join(dirPath, "wp-login.php")) {
		return TypeWordPress
	}
	return TypeGeneric
}

// DocumentRoot returns the web-accessible root directory for the project.
func DocumentRoot(dirPath string, ptype ProjectType) string {
	switch ptype {
	case TypeLaravel:
		return filepath.Join(dirPath, "public")
	case TypeWordPress:
		return dirPath
	default:
		// Generic: prefer public/ if it exists, otherwise the project root.
		pub := filepath.Join(dirPath, "public")
		if info, err := os.Stat(pub); err == nil && info.IsDir() {
			return pub
		}
		return dirPath
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
