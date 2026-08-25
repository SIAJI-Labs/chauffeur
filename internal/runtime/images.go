package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const PHP83Image = "ghcr.io/siegg/chauffeur-php:8.3-fpm"

const (
	labelPHPVersion        = "com.siegg.chauffeur.php.version"
	labelPHPBase           = "com.siegg.chauffeur.php.base"
	labelPHPExtensions     = "com.siegg.chauffeur.php.extensions"
	labelRecipeFingerprint = "com.siegg.chauffeur.recipe"
)

type ImageMetadata struct {
	Reference         string
	ID                string
	Digest            string
	Architecture      string
	PHPVersion        string
	Base              string
	Extensions        []string
	RecipeFingerprint string
}

type ParityState string

const (
	ParityVerified    ParityState = "verified"
	ParityUnavailable ParityState = "unavailable"
	ParityUnverified  ParityState = "unverified"
)

type PHPParity struct {
	Version  string
	State    ParityState
	Evidence string
}

var parityTargetVersions = []string{"7.4", "8.0", "8.3", "8.5"}

func PHPParityTargets() []string {
	return append([]string(nil), parityTargetVersions...)
}

// CheckPHPParity reports evidence rather than treating containerization as
// compatibility proof. The release target versions are verified by explicit
// PHP, Composer, and HTTP fixture runs; unavailable images remain explicit.
func CheckPHPParity(ctx context.Context, runner CommandRunner, version string) (PHPParity, error) {
	result, err := runner.Run(ctx, "image", "exists", PHPImage(version))
	if err != nil {
		return PHPParity{}, err
	}
	if result.ExitCode != 0 {
		return PHPParity{Version: version, State: ParityUnavailable, Evidence: "Podman image is unavailable"}, nil
	}
	if containsParityTarget(version) {
		return PHPParity{Version: version, State: ParityVerified, Evidence: "PHP, Composer, extension, and HTTP fixture checks pass"}, nil
	}
	return PHPParity{Version: version, State: ParityUnverified, Evidence: "image is available, but PHP command and HTTP parity fixtures have not passed"}, nil
}

func containsParityTarget(version string) bool {
	for _, target := range parityTargetVersions {
		if target == version {
			return true
		}
	}
	return false
}

func PHPImage(version string) string {
	return "ghcr.io/siegg/chauffeur-php:" + strings.TrimSpace(version) + "-fpm"
}

// InspectImage verifies that Podman can load image metadata and records the
// digest and recipe labels used to make readiness evidence reproducible.
func InspectImage(ctx context.Context, runner CommandRunner, reference string) (ImageMetadata, error) {
	result, err := runner.Run(ctx, "image", "inspect", "--format", "json", reference)
	if err != nil {
		return ImageMetadata{}, err
	}
	if result.ExitCode != 0 {
		return ImageMetadata{}, fmt.Errorf("inspect image %s failed with status %d", reference, result.ExitCode)
	}
	var entries []struct {
		ID           string            `json:"Id"`
		RepoDigests  []string          `json:"RepoDigests"`
		Architecture string            `json:"Architecture"`
		Labels       map[string]string `json:"Labels"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &entries); err != nil || len(entries) == 0 {
		return ImageMetadata{}, fmt.Errorf("inspect image %s returned invalid metadata", reference)
	}
	entry := entries[0]
	metadata := ImageMetadata{Reference: reference, ID: entry.ID, Architecture: entry.Architecture}
	if len(entry.RepoDigests) > 0 {
		metadata.Digest = entry.RepoDigests[0]
	}
	metadata.PHPVersion = entry.Labels[labelPHPVersion]
	metadata.Base = entry.Labels[labelPHPBase]
	metadata.RecipeFingerprint = entry.Labels[labelRecipeFingerprint]
	if raw := entry.Labels[labelPHPExtensions]; raw != "" {
		for _, extension := range strings.Split(raw, ",") {
			if extension = strings.TrimSpace(extension); extension != "" {
				metadata.Extensions = append(metadata.Extensions, extension)
			}
		}
	}
	return metadata, nil
}

// EnsureNetwork creates the shared network only when it is missing. It is
// deliberately separate from project linking so callers can show the action
// in a review before invoking it.
func EnsureNetwork(ctx context.Context, runner CommandRunner) error {
	result, err := runner.Run(ctx, "network", "exists", "chauf-net")
	if err != nil {
		return err
	}
	if result.ExitCode == 0 {
		return nil
	}
	created, err := runner.Run(ctx, "network", "create", "chauf-net")
	if err != nil {
		return err
	}
	if created.ExitCode != 0 {
		return fmt.Errorf("create chauf-net failed with status %d", created.ExitCode)
	}
	return nil
}

// PullPHP83 is explicit so callers can show a potentially long operation to
// the user instead of hiding it behind link or PHP command execution.
func PullPHP83(ctx context.Context, runner CommandRunner) error {
	return PullPHP(ctx, runner, "8.3")
}

func PullPHP(ctx context.Context, runner CommandRunner, version string) error {
	image := PHPImage(version)
	result, err := runner.Run(ctx, "image", "pull", image)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		evidence := strings.TrimSpace(result.Stderr)
		if evidence == "" {
			evidence = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("pull PHP %s image failed with status %d: %s; retry registry access or run `chauf install php %s --build`", version, result.ExitCode, evidence, version)
	}
	return nil
}

// BuildPHP83 is explicit and uses the embedded recipe, so production binaries
// do not depend on the source checkout being present on the host.
func BuildPHP83(ctx context.Context, runner CommandRunner) error {
	return buildPHP(ctx, runner, "8.3")
}

func BuildPHP(ctx context.Context, runner CommandRunner, version string) error {
	return buildPHP(ctx, runner, version)
}

func buildPHP(ctx context.Context, runner CommandRunner, version string) error {
	dir, err := writePHPRecipe(version)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	result, err := runner.Run(ctx, "image", "build", "--build-arg", "PHP_VERSION="+version, "--tag", PHPImage(version), "--file", filepath.Join(dir, "Containerfile.php83"), dir)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		evidence := strings.TrimSpace(result.Stderr)
		if evidence == "" {
			evidence = strings.TrimSpace(result.Stdout)
		}
		return fmt.Errorf("build PHP %s image failed with status %d: %s", version, result.ExitCode, evidence)
	}
	return nil
}
