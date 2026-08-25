package runtime

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed Containerfile.php83 php-fpm.conf
var phpRecipeFiles embed.FS

func writePHPRecipe(version string) (string, error) {
	dir, err := os.MkdirTemp("", "chauf-php-recipe-")
	if err != nil {
		return "", err
	}
	for _, name := range []string{"Containerfile.php83", "php-fpm.conf"} {
		data, readErr := phpRecipeFiles.ReadFile(name)
		if readErr != nil {
			os.RemoveAll(dir)
			return "", readErr
		}
		if name == "Containerfile.php83" && version != "8.3" {
			recipe := strings.ReplaceAll(string(data), "8.3-fpm-bookworm", version+"-fpm-bookworm")
			if version == "8.0" {
				recipe = strings.ReplaceAll(recipe, version+"-fpm-bookworm", version+"-fpm-bullseye")
			}
			if version == "7.4" {
				recipe = strings.ReplaceAll(recipe, version+"-fpm-bookworm", version+"-fpm-buster")
			}
			recipe = strings.ReplaceAll(recipe, "8.3", version)
			if version == "7.4" {
				recipe = strings.Replace(recipe, "RUN apt-get update \\", "RUN sed -i 's|deb.debian.org|archive.debian.org|g; /security.debian.org/d' /etc/apt/sources.list && apt-get -o Acquire::Check-Valid-Until=false update \\", 1)
			}
			data = []byte(recipe)
		}
		if writeErr := os.WriteFile(filepath.Join(dir, name), data, 0644); writeErr != nil {
			os.RemoveAll(dir)
			return "", writeErr
		}
	}
	return dir, nil
}
