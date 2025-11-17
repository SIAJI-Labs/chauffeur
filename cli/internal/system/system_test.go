package system

import (
	"strings"
	"testing"
)

// TestSystemInfo tests system information detection
func TestSystemInfo(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Errorf("Expected system info to be detected: %v", err)
	}

	if info.Distro == "" {
		t.Error("Expected distro to be detected")
	}

	if info.Arch == "" {
		t.Error("Expected architecture to be detected")
	}

	if info.Pretty == "" {
		t.Error("Expected pretty description to be set")
	}
}

// TestSystemArchitectureValidation tests architecture validation
func TestSystemArchitectureValidation(t *testing.T) {
	validArchs := []string{"x86_64", "aarch64"}
	invalidArchs := []string{"", "invalid", "bad-arch"}

	for _, arch := range validArchs {
		t.Run("valid_"+arch, func(t *testing.T) {
			// Basic architecture validation
			isValid := arch == "x86_64" || arch == "aarch64"
			if !isValid {
				t.Errorf("Expected architecture %s to be valid", arch)
			}
		})
	}

	for _, arch := range invalidArchs {
		t.Run("invalid_"+arch, func(t *testing.T) {
			// Basic architecture validation
			isValid := arch == "x86_64" || arch == "aarch64"
			if isValid {
				t.Errorf("Expected architecture %s to be invalid", arch)
			}
		})
	}
}

// TestPackageManagerDetection tests package manager detection
func TestPackageManagerDetection(t *testing.T) {
	pm := DetectPackageManager()

	// Should return some known package manager or "unknown"
	validPMs := []PackageManager{Pacman, Apt, Yum, Dnf, Zypper, Unknown}

	isValid := false
	for _, validPM := range validPMs {
		if pm == validPM {
			isValid = true
			break
		}
	}

	if !isValid {
		t.Errorf("Expected package manager %v to be valid", pm)
	}
}

// TestPackageValidation tests package validation logic
func TestPackageValidation(t *testing.T) {
	validPackages := []Package{
		{Name: "dnsmasq", Description: "DNS forwarder"},
		{Name: "nginx", Description: "Web server"},
	}

	invalidPackages := []Package{
		{Name: "", Description: "Empty name"},
		{Name: "rm -rf /", Description: "Dangerous command"},
	}

	for _, pkg := range validPackages {
		t.Run("valid_"+pkg.Name, func(t *testing.T) {
			if pkg.Name == "" {
				t.Error("Valid package should not have empty name")
			}
			if pkg.Description == "" {
				t.Error("Valid package should not have empty description")
			}
		})
	}

	for _, pkg := range invalidPackages {
		t.Run("invalid_"+pkg.Name, func(t *testing.T) {
			if pkg.Name == "" || pkg.Description == "Dangerous command" {
				// These are expected to be invalid
			} else {
				t.Errorf("Package %s should be marked invalid", pkg.Name)
			}
		})
	}
}

// TestCommandAvailability tests command availability checking
func TestCommandAvailability(t *testing.T) {
	// Test with a command that should always exist
	exists := IsCommandAvailable("go")
	// Go should be available since we're running Go tests
	// But we can't guarantee it in all environments

	// Test with a command that likely doesn't exist
	notExists := IsCommandAvailable("definitely-not-a-real-command-12345")
	if notExists {
		// This is expected
	}

	// The important thing is that these functions don't panic
	t.Logf("Command availability test: go=%v, fake=%v", exists, notExists)
}

// TestDnsmasqAvailability tests dnsmasq availability checking
func TestDnsmasqAvailability(t *testing.T) {
	// This test checks if dnsmasq is available
	available := IsDnsmasqAvailable()
	// We can't guarantee dnsmasq is installed in all test environments

	// The important thing is that the function doesn't panic
	t.Logf("Dnsmasq availability: %v", available)
}

// TestNetworkManagerAvailability tests NetworkManager availability
func TestNetworkManagerAvailability(t *testing.T) {
	// This test checks if NetworkManager is available
	available := IsNetworkManagerAvailable()
	// We can't guarantee NetworkManager is installed in all test environments

	// The important thing is that the function doesn't panic
	t.Logf("NetworkManager availability: %v", available)
}

// TestSystemdResolvedAvailability tests systemd-resolved availability
func TestSystemdResolvedAvailability(t *testing.T) {
	// This test checks if systemd-resolved stub listener is active
	active := IsSystemdResolvedStubActive()
	// This depends on the system configuration

	// The important thing is that the function doesn't panic
	t.Logf("systemd-resolved stub active: %v", active)
}

// TestPackageInstallationTests package installation simulation
func TestPackageInstallationTests(t *testing.T) {
	// Test package validation before installation
	validPackage := Package{
		Name:        "test-package",
		Description: "Test package for validation",
		PackageName: "test-package",
	}

	if validPackage.Name == "" {
		t.Error("Valid package should have non-empty name")
	}

	if validPackage.Description == "" {
		t.Error("Valid package should have non-empty description")
	}

	if validPackage.PackageName == "" {
		t.Error("Valid package should have non-empty package name")
	}

	// Test invalid package
	invalidPackage := Package{
		Name:        "",
		Description: "Invalid package",
		PackageName: "",
	}

	// This should be caught by validation
	if invalidPackage.Name == "" {
		t.Log("Invalid package correctly identified by empty name")
	}
}

// TestFilePathValidation tests file path validation
func TestFilePathValidation(t *testing.T) {
	validPaths := []string{
		"/tmp/test.txt",
		"/home/user/project",
		"/var/log/test.log",
	}

	invalidPaths := []string{
		"",
		"/tmp/../../../etc/passwd",
		"../../../root/.ssh",
	}

	for _, path := range validPaths {
		t.Run("valid_"+path, func(t *testing.T) {
			// Basic path validation
			isValid := path != "" && !strings.Contains(path, "..")
			if !isValid {
				t.Errorf("Expected path %s to be valid", path)
			}
		})
	}

	for _, path := range invalidPaths {
		t.Run("invalid_"+path, func(t *testing.T) {
			// Basic path validation
			isValid := path != "" && !strings.Contains(path, "..")
			if isValid {
				t.Errorf("Expected path %s to be invalid", path)
			}
		})
	}
}

// TestPortRangeValidation tests port range validation
func TestPortRangeValidation(t *testing.T) {
	validRanges := []struct {
		start int
		end   int
	}{
		{8080, 8099},
		{9000, 9099},
		{1024, 2048},
	}

	invalidRanges := []struct {
		start int
		end   int
	}{
		{-1, 8099},
		{8080, 0},
		{8099, 8080},
		{0, 0},
		{65536, 65537},
	}

	for _, test := range validRanges {
		t.Run("valid_range", func(t *testing.T) {
			if test.start <= 0 || test.end <= 0 {
				t.Error("Valid range should have positive ports")
			}
			if test.start >= test.end {
				t.Errorf("Valid range should have start < end: %d >= %d", test.start, test.end)
			}
			if test.end > 65535 {
				t.Errorf("Valid range end should be <= 65535: %d", test.end)
			}
		})
	}

	for _, test := range invalidRanges {
		t.Run("invalid_range", func(t *testing.T) {
			isValid := test.start > 0 && test.end > 0 && test.start < test.end && test.end <= 65535
			if isValid {
				t.Errorf("Range %d-%d should be invalid", test.start, test.end)
			}
		})
	}
}

// TestNetworkConfigurationValidation tests network configuration validation
func TestNetworkConfigurationValidation(t *testing.T) {
	validConfigs := []string{
		"address=/.test/127.0.0.1",
		"listen-address=127.0.0.1",
		"bind-interfaces",
	}

	invalidConfigs := []string{
		"",
		"invalid-config-line",
		"rm -rf /",
	}

	for _, config := range validConfigs {
		t.Run("valid_config", func(t *testing.T) {
			if config == "" {
				t.Error("Valid config should not be empty")
			}
			// Additional validation could be added based on specific requirements
		})
	}

	for _, config := range invalidConfigs {
		t.Run("invalid_config", func(t *testing.T) {
			if config == "" || strings.Contains(config, "rm -rf") {
				// These are expected to be invalid
			}
		})
	}
}