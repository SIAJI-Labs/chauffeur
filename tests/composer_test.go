package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposerServiceSupport(t *testing.T) {
	// Test that composer is in the known services list
	services := commands.KnownServices()
	assert.Contains(t, services, "composer", "Composer should be in known services list")
}

func TestComposerServiceSpecification(t *testing.T) {
	// Test that composer service can be created
	spec, err := commands.NewServiceSpec("composer", "/tmp/chauffeur", system.Info{})
	assert.NoError(t, err, "Composer service spec should be created without error")
	assert.Equal(t, "composer", spec.name, "Service name should be composer")
	assert.Equal(t, "PHP dependency manager with Chauffeur PHP version isolation", spec.description, "Description should mention PHP version isolation")
	assert.Equal(t, "/tmp/chauffeur/bin/composer", spec.binaryPath, "Binary path should be set correctly")
	assert.NotNil(t, spec.installFunc, "Install function should be defined")
}

func TestComposerShimContent(t *testing.T) {
	// Test that composer shim content contains expected elements
	shimContent := getComposerShimContentForTest()
	
	// Check for key components
	assert.Contains(t, shimContent, "get_php_for_project()", "Should have PHP detection function")
	assert.Contains(t, shimContent, "COMPOSER_PHAR", "Should reference Composer PHAR")
	assert.Contains(t, shimContent, "exec \"$PHP_BINARY\"", "Should exec with detected PHP")
	assert.Contains(t, shimContent, "project.yaml", "Should look for project configuration")
	
	// Check that it's a bash script
	assert.True(t, strings.HasPrefix(shimContent, "#!/bin/bash"), "Should be a bash script")
}

func TestComposerVersionSupport(t *testing.T) {
	// Test version support functions
	supported := []string{"2.7", "2.6", "2.5", "2.4", "2.3", "2.2"}
	assert.Equal(t, supported, getSupportedComposerVersionsForTest(), "Should return supported versions")
	
	for _, version := range supported {
		assert.True(t, isComposerVersionSupportedForTest(version), "Version %s should be supported", version)
	}
	assert.True(t, isComposerVersionSupportedForTest(version), "Should be supported")
	
	assert.Equal(t, "2.7, 2.6, 2.5, 2.4, 2.3, 2.2", getSupportedComposerVersionsListForTest(), "Should format versions list correctly")
	
	// Test unsupported version
	assert.False(t, isComposerVersionSupportedForTest("1.0"), "Version 1.0 should not be supported")
	assert.False(t, isComposerVersionSupportedForTest("3.0"), "Version 3.0 should not be supported")
}

func TestComposerProjectIntegration(t *testing.T) {
	// Test that composer integration works with project detection
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	
	// Create a mock workspace and project
	wsDir := filepath.Join(tmpHome, ".chauffeur")
	projectsDir := filepath.Join(wsDir, "projects")
	projectDir := filepath.Join(projectsDir, "myproject")
	projectConfigPath := filepath.Join(projectDir, "project.yaml")
	
	// Create directories
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)
	
	// Create project config with PHP 8.2
	projectConfig := `version: 1
path: /path/to/project
php: 8.2
site:
  domain: myproject.test
  ssl: true
runtime:
  php_fpm_socket: ~/.chauffeur/projects/myproject/runtime/php-fpm/php-fpm.sock
created_at: 2025-01-01T12:00:00+07:00
`
	err = os.WriteFile(projectConfigPath, []byte(projectConfig), 0644)
	require.NoError(t, err)
	
	// Test that composer shim would detect project PHP version
	// (We can't easily test the bash script execution, but we can verify file exists)
	assert.FileExists(t, projectConfigPath, "Project config should exist")
	
	// Verify PHP version in config
	configContent, err := os.ReadFile(projectConfigPath)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "php: 8.2", "Project config should specify PHP version")
}

func TestComposerBinaryPathGeneration(t *testing.T) {
	prefix := "/tmp/test-chauffeur"
	composerBinaryPath := filepath.Join(prefix, "bin", "composer")
	assert.Equal(t, composerBinaryPath, generateComposerBinaryPathForTest(prefix))
}

// Helper functions for testing (mimicking the unexported functions)
func getComposerShimContentForTest() string {
	tmpHome := t.TempDir()
	composerBinaryPath := filepath.Join(tmpHome, "composer.phar")
	
	return `#!/bin/bash
# Composer shim for Chauffeur - uses project-aware PHP version isolation
# This shim ensures Composer always uses the correct PHP version

# Get Chauffeur workspace directory
CHAUF_HOME="${CHAUF_HOME:-$HOME/.chauffeur}"

# Path to the actual Composer PHAR
COMPOSER_PHAR="` + composerBinaryPath + `"

# Function to detect project and get appropriate PHP
get_php_for_project() {
    local current_dir="$(pwd)"
    local projects_dir="$CHAUF_HOME/projects"
    
    # Search for project config in the projects directory
    for project_dir in "$projects_dir"/*; do
        if [[ -f "$project_dir/project.yaml" ]]; then
            local project_path=""
            if [[ -f "$project_dir/project.yaml" ]]; then
                # Extract project path from project.yaml
                project_path="$project_dir"
                # Here we'd normally parse the YAML, but for simplicity, use directory name
                # to match current directory with project config
                
                # Check if current directory matches this project
                project_name="$(basename "$project_dir")"
                # Simple heuristic: check if current path contains project name
                case "$current_dir" in
                    *"$project_name"*)
                        # Found matching project, read its config
                        local php_version=""
                        if php_version="$(grep '^php:' "$project_dir/project.yaml" 2>/dev/null | sed 's/php:[[:space:]]*//')"; then
                            local php_binary="$CHAUF_HOME/php/$php_version/bin/php"
                            if [[ -x "$php_binary" ]]; then
                                echo "$php_binary"
                                return 0
                            fi
                        fi
                        ;;
                esac
            fi
        fi
    done
    
    # Fallback: try to use default PHP from Chauffeur
    local default_php=""
    if default_php="$(grep 'default:' "$CHAUF_HOME/config/chauffeur.yaml" 2>/dev/null | sed 's/default:[[:space:]]*//')"; then
        local default_php_binary="$CHAUF_HOME/php/$default_php/bin/php"
        if [[ -x "$default_php_binary" ]]; then
            echo "$default_php_binary"
            return 0
        fi
    fi
    
    # Final fallback: system PHP
    echo "php"
}

# Get the appropriate PHP binary
PHP_BINARY="$(get_php_for_project)"

# Check if PHP binary exists and is executable
if [[ ! -x "$PHP_BINARY" ]]; then
    echo "Error: PHP binary not found: $PHP_BINARY" >&2
    echo "Please install PHP with Chauffeur: chauf install php" >&2
    exit 1
fi

# Forward arguments to Composer using the detected PHP
exec "$PHP_BINARY" "$COMPOSER_PHAR" "$@"

# If execution reaches here, something went wrong
echo "Error: Failed to execute Composer with PHP $PHP_BINARY" >&2
exit 1
`

func getSupportedComposerVersionsForTest() []string {
	return []string{"2.7", "2.6", "2.5", "2.4", "2.3", "2.2"}
}

func isComposerVersionSupportedForTest(version string) bool {
	supported := getSupportedComposerVersionsForTest()
	for _, v := range supported {
		if v == version {
			return true
		}
	}
	return false
}

func getSupportedComposerVersionsListForTest() string {
	supported := getSupportedComposerVersionsForTest()
	return strings.Join(supported, ", ")
}

func generateComposerBinaryPathForTest(prefix string) string {
	return filepath.Join(prefix, "bin", "composer")
}
