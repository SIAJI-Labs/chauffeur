package installers

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallOptions tests InstallOptions structure
func TestInstallOptions(t *testing.T) {
	options := InstallOptions{
		Prefix:    "/test/.chauffeur",
		Force:     true,
		Client:    &http.Client{},
		EnableGD:  false,
	}

	if options.Prefix != "/test/.chauffeur" {
		t.Errorf("Expected prefix '/test/.chauffeur', got '%s'", options.Prefix)
	}

	if !options.Force {
		t.Error("Expected Force to be true")
	}

	if options.Client == nil {
		t.Error("Expected Client to be set")
	}

	if options.EnableGD {
		t.Error("Expected EnableGD to be false")
	}
}

// TestInstallDefaults tests default options
func TestInstallDefaults(t *testing.T) {
	options := InstallOptions{}

	// Test zero values
	if options.Prefix != "" {
		t.Errorf("Expected empty prefix, got '%s'", options.Prefix)
	}

	if options.Force {
		t.Error("Expected Force to be false by default")
	}

	if options.EnableGD {
		t.Error("Expected EnableGD to be false by default")
	}
}

// TestShimContent tests shim content generation
func TestShimContent(t *testing.T) {
	testCases := []struct {
		name   string
		target string
		expect string
	}{
		{
			name:   "nginx",
			target: "/path/to/nginx",
			expect: "#!/usr/bin/env bash",
		},
		{
			name:   "php",
			target: "/path/to/php",
			expect: "Get workspace directory",
		},
		{
			name:   "custom-binary",
			target: "/usr/bin/custom",
			expect: "TARGET=\"/usr/bin/custom\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := shimContent(tc.name, tc.target)
			if content == "" {
				t.Error("Expected shim content to be generated")
			}

			if !strings.Contains(content, tc.expect) {
				t.Errorf("Expected shim content to contain '%s'", tc.expect)
			}

			// All shims should start with shebang
			if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
				t.Error("Expected shim to start with bash shebang")
			}
		})
	}
}

// TestProjectAwarePHPShimContent tests PHP shim content generation
func TestProjectAwarePHPShimContent(t *testing.T) {
	content := ProjectAwarePHPShimContent()

	if content == "" {
		t.Error("Expected PHP shim content to be generated")
	}

	expectedParts := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"WORKSPACE=\"${HOME}/.chauffeur\"",
		"find_project_config",
		"PROJECT_CONFIG",
		"PHP_BINARY",
		"exec \"$PHP_BINARY\"",
	}

	for _, part := range expectedParts {
		if !strings.Contains(content, part) {
			t.Errorf("Expected PHP shim content to contain '%s'", part)
		}
	}
}

// TestWriteShim tests shim writing functionality
func TestWriteShim(t *testing.T) {
	tempDir := t.TempDir()
	prefix := filepath.Join(tempDir, ".chauffeur")

	// Test successful shim creation
	err := writeShim(prefix, "test-binary", "/path/to/binary")
	if err != nil {
		t.Errorf("Expected writeShim to succeed: %v", err)
	}

	// Verify shim file exists
	shimPath := filepath.Join(prefix, "bin", "test-binary")
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Errorf("Expected shim file to exist at %s", shimPath)
	}

	// Verify shim is executable
	info, err := os.Stat(shimPath)
	if err != nil {
		t.Errorf("Expected to stat shim file: %v", err)
	}

	if info.Mode().Perm()&0o111 == 0 {
		t.Error("Expected shim file to be executable")
	}

	// Verify shim content
	content, err := os.ReadFile(shimPath)
	if err != nil {
		t.Errorf("Expected to read shim file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected shim file to have content")
	}
}

// TestWriteShimErrorHandling tests shim write error handling
func TestWriteShimErrorHandling(t *testing.T) {
	// Test with invalid prefix (should fail)
	err := writeShim("/invalid/path/that/cannot/be/created", "test", "/target")
	if err == nil {
		t.Error("Expected writeShim to fail with invalid prefix")
	}
}

// TestValidateChecksum tests checksum validation
func TestValidateChecksum(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	// Write test file
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Calculate expected checksums
	sha256Hasher := sha256.New()
	sha256Hasher.Write([]byte(testContent))
	expectedSHA256 := hex.EncodeToString(sha256Hasher.Sum(nil))

	sha512Hasher := sha512.New()
	sha512Hasher.Write([]byte(testContent))
	expectedSHA512 := hex.EncodeToString(sha512Hasher.Sum(nil))

	testCases := []struct {
		name     string
		path     string
		expected string
		shouldErr bool
	}{
		{
			name:      "valid sha256",
			path:      testFile,
			expected:  expectedSHA256,
			shouldErr: false,
		},
		{
			name:      "valid sha512",
			path:      testFile,
			expected:  expectedSHA512,
			shouldErr: false,
		},
		{
			name:      "invalid checksum",
			path:      testFile,
			expected:  "invalidchecksum123",
			shouldErr: true,
		},
		{
			name:      "empty checksum",
			path:      testFile,
			expected:  "",
			shouldErr: true,
		},
		{
			name:      "nonexistent file",
			path:      filepath.Join(tempDir, "nonexistent.txt"),
			expected:  expectedSHA256,
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateChecksum(tc.path, tc.expected)
			if tc.shouldErr && err == nil {
				t.Errorf("Expected validateChecksum to fail for '%s'", tc.name)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected validateChecksum to succeed for '%s': %v", tc.name, err)
			}
		})
	}
}

// TestChecksumNormalization tests checksum format normalization
func TestChecksumNormalization(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Test content"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Calculate correct checksum
	hasher := sha256.New()
	hasher.Write([]byte(testContent))
	correctChecksum := hex.EncodeToString(hasher.Sum(nil))

	testCases := []struct {
		name     string
		checksum string
		shouldErr bool
	}{
		{
			name:      "plain checksum",
			checksum:  correctChecksum,
			shouldErr: false,
		},
		{
			name:      "checksum with filename",
			checksum:  fmt.Sprintf("%s  test.txt", correctChecksum),
			shouldErr: false,
		},
		{
			name:      "checksum with prefix",
			checksum:  fmt.Sprintf("sha256:%s", correctChecksum),
			shouldErr: false,
		},
		{
			name:      "checksum with whitespace",
			checksum:  fmt.Sprintf("  %s  ", correctChecksum),
			shouldErr: false,
		},
		{
			name:      "unsupported length",
			checksum:  "12345",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateChecksum(testFile, tc.checksum)
			if tc.shouldErr && err == nil {
				t.Errorf("Expected validation to fail for checksum format '%s'", tc.name)
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected validation to succeed for checksum format '%s': %v", tc.name, err)
			}
		})
	}
}

// TestFileHashFunctions tests file hashing functions
func TestFileHashFunctions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hash this content!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test SHA256
	sha256Hash, err := fileSHA256(testFile)
	if err != nil {
		t.Errorf("Expected fileSHA256 to succeed: %v", err)
	}
	if len(sha256Hash) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("Expected SHA256 hash to be 64 characters, got %d", len(sha256Hash))
	}

	// Test SHA512
	sha512Hash, err := fileSHA512(testFile)
	if err != nil {
		t.Errorf("Expected fileSHA512 to succeed: %v", err)
	}
	if len(sha512Hash) != 128 { // SHA512 produces 128 hex characters
		t.Errorf("Expected SHA512 hash to be 128 characters, got %d", len(sha512Hash))
	}

	// Verify hashes are different (they should be)
	if sha256Hash == sha512Hash {
		t.Error("Expected SHA256 and SHA512 hashes to be different")
	}
}

// TestFileHashErrorHandling tests file hash error handling
func TestFileHashErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	nonexistentFile := filepath.Join(tempDir, "nonexistent.txt")

	// Test SHA256 with nonexistent file
	_, err := fileSHA256(nonexistentFile)
	if err == nil {
		t.Error("Expected fileSHA256 to fail with nonexistent file")
	}

	// Test SHA512 with nonexistent file
	_, err = fileSHA512(nonexistentFile)
	if err == nil {
		t.Error("Expected fileSHA512 to fail with nonexistent file")
	}
}

// TestCommandError tests command error functionality
func TestCommandError(t *testing.T) {
	cmdErr := commandError{
		Name:   "test-command",
		Args:   []string{"--flag", "value"},
		Err:    fmt.Errorf("command failed"),
		Stdout: "output line 1",
		Stderr: "error line 1",
	}

	// Test Error() method
	errorMsg := cmdErr.Error()
	expectedParts := []string{"test-command", "--flag", "value", "command failed"}
	for _, part := range expectedParts {
		if !strings.Contains(errorMsg, part) {
			t.Errorf("Expected error message to contain '%s'", part)
		}
	}

	// Test Detail() method
	detailMsg := cmdErr.Detail()
	expectedDetailParts := []string{
		"test-command --flag value failed",
		"command failed",
		"stdout:",
		"output line 1",
		"stderr:",
		"error line 1",
	}
	for _, part := range expectedDetailParts {
		if !strings.Contains(detailMsg, part) {
			t.Errorf("Expected detail message to contain '%s'", part)
		}
	}

	// Test with empty fields
	emptyErr := commandError{
		Name: "empty-test",
	}

	if emptyErr.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	if emptyErr.Detail() == "" {
		t.Error("Expected non-empty detail message")
	}
}

// TestDownloadToFile tests download functionality (without actual network calls)
func TestDownloadToFile(t *testing.T) {
	client := &http.Client{}

	// Test that the function exists and doesn't panic with basic parameters
	// We can't easily test actual downloads without mocking or network calls
	_, err := downloadToFile(client, "http://example.com", "/tmp/test", "Test Label")

	// This will likely fail due to network/file issues, but should not panic
	if err != nil {
		t.Logf("downloadToFile returned expected error: %v", err)
	}
}

// TestDownloadText tests text download functionality
func TestDownloadText(t *testing.T) {
	client := &http.Client{}

	// Test that the function exists and doesn't panic
	_, err := downloadText(client, "http://example.com")

	// This will likely fail due to network issues, but should not panic
	if err != nil {
		t.Logf("downloadText returned expected error: %v", err)
	}
}

// TestChecksumHelpers tests checksum helper functions
func TestChecksumHelpers(t *testing.T) {
	client := &http.Client{}

	// Test checksumFromList
	_, err := checksumFromList(client, "http://example.com/checksums.txt", "test.txt")
	if err != nil {
		t.Logf("checksumFromList returned expected error: %v", err)
	}

	// Test checksumFromContent
	checksumContent := `abc123def456...  test.txt
def789abc012...  other.txt
`
	checksum, err := checksumFromContent(checksumContent, "test.txt")
	if err != nil {
		t.Errorf("Expected checksumFromContent to succeed: %v", err)
	}
	if checksum != "abc123def456..." {
		t.Errorf("Expected checksum 'abc123def456...', got '%s'", checksum)
	}

	// Test with nonexistent file
	_, err = checksumFromContent(checksumContent, "nonexistent.txt")
	if err == nil {
		t.Error("Expected checksumFromContent to fail for nonexistent file")
	}
}

// TestShimPermissions tests that shims are created with correct permissions
func TestShimPermissions(t *testing.T) {
	tempDir := t.TempDir()
	prefix := filepath.Join(tempDir, ".chauffeur")

	err := writeShim(prefix, "test-shim", "/path/to/binary")
	if err != nil {
		t.Fatalf("Failed to write shim: %v", err)
	}

	shimPath := filepath.Join(prefix, "bin", "test-shim")

	// Check file exists and is executable
	info, err := os.Stat(shimPath)
	if err != nil {
		t.Fatalf("Failed to stat shim: %v", err)
	}

	// Check that owner has execute permission
	if info.Mode().Perm()&0o100 == 0 {
		t.Error("Expected shim to have owner execute permission")
	}

	// Check that group has execute permission
	if info.Mode().Perm()&0o010 == 0 {
		t.Error("Expected shim to have group execute permission")
	}

	// Check that others have execute permission
	if info.Mode().Perm()&0o001 == 0 {
		t.Error("Expected shim to have others execute permission")
	}
}

// TestMultipleShims tests creating multiple shims
func TestMultipleShims(t *testing.T) {
	tempDir := t.TempDir()
	prefix := filepath.Join(tempDir, ".chauffeur")

	shims := map[string]string{
		"nginx": "/path/to/nginx",
		"php":   "/path/to/php",
		"node":  "/path/to/node",
	}

	for name, target := range shims {
		err := writeShim(prefix, name, target)
		if err != nil {
			t.Errorf("Failed to write shim %s: %v", name, err)
		}

		// Verify shim exists
		shimPath := filepath.Join(prefix, "bin", name)
		if _, err := os.Stat(shimPath); os.IsNotExist(err) {
			t.Errorf("Expected shim %s to exist", name)
		}
	}

	// Verify bin directory was created
	binDir := filepath.Join(prefix, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		t.Error("Expected bin directory to exist")
	}
}

// TestLegacyDependencyValidation tests legacy PHP dependency validation
func TestLegacyDependencyValidation(t *testing.T) {
	testCases := []struct {
		name           string
		phpVersion     string
		pkgName        string
		detectedVer    string
		shouldErr      bool
		expectedErrMsg string
	}{
		{
			name:        "modern PHP with normal version",
			phpVersion:  "8.3",
			pkgName:     "libxml-2.0",
			detectedVer: "2.12.0",
			shouldErr:   false,
		},
		{
			name:        "legacy PHP with compatible version",
			phpVersion:  "7.4",
			pkgName:     "libxml-2.0",
			detectedVer: "2.11.0",
			shouldErr:   false,
		},
		{
			name:           "legacy PHP with too new version",
			phpVersion:     "7.4",
			pkgName:        "libxml-2.0",
			detectedVer:    "2.13.0",
			shouldErr:      true,
			expectedErrMsg: "version too new for legacy PHP 7.4",
		},
		{
			name:        "legacy PHP with no constraints",
			phpVersion:  "7.4",
			pkgName:     "libcurl",
			detectedVer: "7.80.0",
			shouldErr:   false,
		},
		{
			name:           "legacy PHP with ImageMagick too new",
			phpVersion:     "7.4",
			pkgName:        "MagickWand",
			detectedVer:    "7.2.0",
			shouldErr:      true,
			expectedErrMsg: "version too new for legacy PHP 7.4",
		},
		{
			name:        "no PHP version specified",
			phpVersion:  "",
			pkgName:     "libxml-2.0",
			detectedVer: "2.13.0",
			shouldErr:   false,
		},
		{
			name:        "non-legacy PHP version",
			phpVersion:  "8.2",
			pkgName:     "libxml-2.0",
			detectedVer: "2.13.0",
			shouldErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We'll mock the ensurePkgRequirement call by simulating its success
			// and testing the legacy validation logic directly
			legacyReq, hasLegacyReq := getLegacyDependencyRequirement(tc.phpVersion, tc.pkgName)

			// If there's a legacy requirement, test the validation logic
			if hasLegacyReq && tc.phpVersion != "" && isLegacyPHPVersion(tc.phpVersion) {
				// Test max version constraint
				if legacyReq.MaxVersion != "" && compareSemver(tc.detectedVer, legacyReq.MaxVersion) > 0 {
					if !tc.shouldErr {
						t.Errorf("Expected legacy constraint to pass for %s %s vs %s", tc.pkgName, tc.detectedVer, legacyReq.MaxVersion)
					}
					if !strings.Contains(tc.expectedErrMsg, "too new") {
						t.Errorf("Expected 'too new' error message for version constraint violation")
					}
				}

				// Test min version constraint
				if legacyReq.MinVersion != "" && compareSemver(tc.detectedVer, legacyReq.MinVersion) < 0 {
					if !tc.shouldErr {
						t.Errorf("Expected legacy constraint to pass for %s %s vs %s", tc.pkgName, tc.detectedVer, legacyReq.MinVersion)
					}
					if !strings.Contains(tc.expectedErrMsg, "too old") {
						t.Errorf("Expected 'too old' error message for version constraint violation")
					}
				}
			} else if tc.shouldErr {
				t.Errorf("Expected test case '%s' to fail but no legacy constraints found", tc.name)
			}
		})
	}
}

// TestLegacyVersionDetection tests legacy PHP version detection
func TestLegacyVersionDetection(t *testing.T) {
	testCases := []struct {
		version   string
		isLegacy  bool
	}{
		{"7.4", true},
		{"8.0", true},
		{"8.1", false},
		{"8.2", false},
		{"8.3", false},
		{"8.4", false},
		{"", false},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("PHP %s", tc.version), func(t *testing.T) {
			result := isLegacyPHPVersion(tc.version)
			if result != tc.isLegacy {
				t.Errorf("Expected isLegacyPHPVersion(%s) to be %v, got %v", tc.version, tc.isLegacy, result)
			}
		})
	}
}

// TestLegacyDependencyMatrix tests the legacy dependency matrix structure
func TestLegacyDependencyMatrix(t *testing.T) {
	// Test that the matrix contains expected PHP versions
	expectedVersions := []string{"7.4", "8.0"}
	foundVersions := make(map[string]bool)

	for _, dep := range legacyDependencyMatrix {
		foundVersions[dep.PHPVersion] = true

		// Validate structure
		if dep.PHPVersion == "" {
			t.Error("Expected PHPVersion to be set in legacy dependency matrix")
		}
		if dep.PackageName == "" {
			t.Error("Expected PackageName to be set in legacy dependency matrix")
		}
	}

	for _, version := range expectedVersions {
		if !foundVersions[version] {
			t.Errorf("Expected legacy dependency matrix to contain constraints for PHP %s", version)
		}
	}
}

// TestCompareSemver tests semantic version comparison
func TestCompareSemver(t *testing.T) {
	testCases := []struct {
		a        string
		b        string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.9.9", "2.0.0", -1},
		{"1.2.0", "1.2", 0}, // Should handle missing patch version
		{"1.2", "1.2.0", 0},  // Should handle missing patch version
		{"1.10.0", "1.9.0", 1}, // Should handle double-digit versions
		{"1.0.0", "1.0", 0},  // Should handle missing patch version
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s vs %s", tc.a, tc.b), func(t *testing.T) {
			result := compareSemver(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Expected compareSemver(%s, %s) to be %d, got %d", tc.a, tc.b, tc.expected, result)
			}
		})
	}
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	// Test empty name for shim
	content := shimContent("", "/path/to/binary")
	if content == "" {
		t.Error("Expected shim content even with empty name")
	}

	// Test empty target for shim
	content = shimContent("test", "")
	if !strings.Contains(content, "TARGET=\"\"") {
		t.Error("Expected shim content to contain empty target")
	}

	// Test validateChecksum with empty path
	err := validateChecksum("", "abc123")
	if err == nil {
		t.Error("Expected validateChecksum to fail with empty path")
	}
}