package installers

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/siaji/chauffeur/cli/lib"
)

/**
 * writeShim ensures the shim executable points to the provided target binary.
 *
 * @param prefix Workspace root where shims live.
 * @param name   Shim filename to create or update.
 * @param target Absolute path to the managed binary.
 * @return error when the shim cannot be written.
 */
func writeShim(prefix, name, target string) error {
	binDir := filepath.Join(prefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("ensure bin dir: %w", err)
	}

	scriptPath := filepath.Join(binDir, name)
	if err := os.WriteFile(scriptPath, []byte(shimContent(name, target)), 0o755); err != nil {
		return fmt.Errorf("write shim %s: %w", scriptPath, err)
	}
	return nil
}

/**
 * shimContent constructs the shell script content for a shim.
 *
 * @param name   Shim executable name.
 * @param target Absolute path to the real binary.
 * @return Shell script body that delegates to target.
 */
func shimContent(name, target string) string {
	if name == "php" {
		return ProjectAwarePHPShimContent()
	}
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
TARGET="%s"
if [[ ! -x "$TARGET" ]]; then
  echo "%s binary is missing at $TARGET" >&2
  exit 1
fi
exec "$TARGET" "$@"
`, target, name)
}

// ProjectAwarePHPShimContent generates content for a PHP shim that checks for project isolation.
func ProjectAwarePHPShimContent() string {
	return `#!/usr/bin/env bash
set -euo pipefail

# Get workspace directory
WORKSPACE="${HOME}/.chauffeur"
if [[ ! -d "$WORKSPACE" ]]; then
  echo "Chauffeur workspace not found at $WORKSPACE" >&2
  exit 1
fi

# Get current directory
CWD="$(pwd)"

# Function to find project config
find_project_config() {
  local project_dir="$1"
  local cwd="$2"
  
  # First try standard projects directory approach
  local projects_dir="$project_dir/projects"
  local slug="$(basename "$cwd")"
  local project_config="$projects_dir/$slug/project.yaml"
  if [[ -f "$project_config" ]]; then
    echo "$project_config"
    return 0
  fi
  
  # Additional fallback: search all project configs for matching path
  if [[ -d "$projects_dir" ]]; then
    for config_file in "$projects_dir"/*/project.yaml; do
      if [[ -f "$config_file" ]]; then
        # Extract path from config and check if it matches current directory or parent
        local config_path
        config_path=$(grep "^path:" "$config_file" 2>/dev/null | sed 's/path: *//' | xargs)
        if [[ -n "$config_path" ]]; then
          # Check if current directory is the project directory or a subdirectory
          if [[ "$cwd" == "$config_path"* ]]; then
            echo "$config_file"
            return 0
          fi
        fi
      fi
    done
  fi
  
  return 1
}

# Find project configuration
PROJECT_CONFIG=""
PHP_VERSION=""
if PROJECT_CONFIG=$(find_project_config "$WORKSPACE" "$CWD"); then
  # Extract PHP version from project config
  if grep -q "^php:" "$PROJECT_CONFIG" 2>/dev/null; then
    PHP_VERSION=$(grep "^php:" "$PROJECT_CONFIG" | sed 's/.*php: *//' | xargs)
  fi
fi

# Determine PHP binary path
if [[ -n "$PHP_VERSION" ]]; then
  # Use isolated PHP version
  PHP_BINARY="$WORKSPACE/php/$PHP_VERSION/bin/php"
else
  # Use default PHP version (check config, then fallback to 8.3)
  if [[ -f "$WORKSPACE/config/chauffeur.yaml" ]]; then
    DEFAULT_VERSION="$(grep "default:" "$WORKSPACE/config/chauffeur.yaml" | sed 's/.*default: *//' | xargs)"
    PHP_BINARY="$WORKSPACE/php/$DEFAULT_VERSION/bin/php"
  else
    # Fallback to PHP 8.3 if no config exists
    PHP_BINARY="$WORKSPACE/php/8.3/bin/php"
  fi
fi

if [[ ! -x "$PHP_BINARY" ]]; then
  echo "PHP binary not found at $PHP_BINARY" >&2
  echo "Please run 'chauf install php <version>' first" >&2
  exit 1
fi

exec "$PHP_BINARY" "$@"
`
}

/**
 * downloadToFile streams a remote file down to dest and returns the byte count.
 * When label is provided, a simple progress bar is rendered to stdout.
 *
 * @param client HTTP client used for the request.
 * @param url    Remote resource to download.
 * @param dest   Local path to persist the file.
 * @param label  Optional label for progress rendering; leave empty to disable.
 * @return Number of bytes written and an error, if any.
 */
func downloadToFile(client *http.Client, url, dest, label string) (int64, error) {
	return lib.DownloadToFile(client, url, dest, label)
}

/**
 * downloadText fetches the resource at url and returns it as trimmed text.
 *
 * @param client HTTP client used for the request.
 * @param url    Remote resource to download.
 * @return Trimmed textual contents of the response body.
 */
func downloadText(client *http.Client, url string) (string, error) {
	return lib.DownloadText(client, url)
}

/**
 * checksumFromList downloads a checksum list and extracts the value for assetName.
 *
 * @param client HTTP client used to fetch the list.
 * @param url    Location of the checksum manifest.
 * @param assetName File name whose checksum is required.
 * @return Matching checksum string for the asset.
 */
func checksumFromList(client *http.Client, url, assetName string) (string, error) {
	return lib.ChecksumFromList(client, url, assetName)
}

/**
 * checksumFromContent parses checksum text and returns the hash for assetName.
 *
 * @param content  Raw checksum list contents.
 * @param assetName File name to match against the list.
 * @return Checksum string matching assetName.
 */
func checksumFromContent(content, assetName string) (string, error) {
	return lib.ChecksumFromContent(content, assetName)
}

/**
 * validateChecksum reads the file at path and verifies the digest matches expected.
 *
 * @param path     Local file path to hash.
 * @param expected Expected checksum string (sha256 or sha512).
 * @return error when the calculated digest differs from expected.
 */
func validateChecksum(path, expected string) error {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fmt.Errorf("empty checksum provided")
	}

	// Split away filenames or prefixes like "sha256:"
	if strings.Contains(expected, " ") {
		expected = strings.Fields(expected)[0]
	}
	if strings.Contains(expected, ":") {
		parts := strings.Split(expected, ":")
		expected = parts[len(parts)-1]
	}

	expected = strings.TrimSpace(expected)
	if expected == "" {
		return fmt.Errorf("empty checksum after normalization")
	}

	var (
		hasher hash.Hash
		err    error
	)

	switch len(expected) {
	case 64:
		hasher = sha256.New()
	case 128:
		hasher = sha512.New()
	default:
		return fmt.Errorf("unsupported checksum length %d", len(expected))
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch: expected %s got %s", expected, actual)
	}

	return nil
}

func fileSHA256(path string) (string, error) {
	return computeFileHash(path, sha256.New())
}

func fileSHA512(path string) (string, error) {
	return computeFileHash(path, sha512.New())
}

func computeFileHash(path string, hasher hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}


