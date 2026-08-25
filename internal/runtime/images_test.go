package runtime

import (
	"context"
	"strings"
	"testing"
)

func TestEnsureNetworkIsIdempotent(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{ExitCode: 0}}
	if err := EnsureNetwork(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	if len(runner.args) != 1 || runner.args[0][0] != "network" || runner.args[0][1] != "exists" {
		t.Fatalf("unexpected calls: %#v", runner.args)
	}
}

func TestPullPHP83UsesPinnedFriendlyImageReference(t *testing.T) {
	runner := &recordingRunner{}
	if err := PullPHP83(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	want := []string{"image", "pull", PHP83Image}
	for i := range want {
		if runner.args[0][i] != want[i] {
			t.Fatalf("args = %#v, want %#v", runner.args, want)
		}
	}
}

func TestCheckPHPParityDistinguishesEvidenceStates(t *testing.T) {
	runner := &sequenceRunner{results: []CommandResult{{ExitCode: 1}, {ExitCode: 0}, {ExitCode: 0}}}
	unavailable, err := CheckPHPParity(context.Background(), runner, "7.4")
	if err != nil || unavailable.State != ParityUnavailable {
		t.Fatalf("unavailable = %+v, err = %v", unavailable, err)
	}
	available, err := CheckPHPParity(context.Background(), runner, "8.0")
	if err != nil || available.State != ParityVerified {
		t.Fatalf("available = %+v, err = %v", available, err)
	}
	verified, err := CheckPHPParity(context.Background(), runner, "8.5")
	if err != nil || verified.State != ParityVerified {
		t.Fatalf("verified = %+v, err = %v", verified, err)
	}
}

func TestPHPParityTargetsIncludeReleaseGateVersions(t *testing.T) {
	want := []string{"7.4", "8.0", "8.3", "8.5"}
	got := PHPParityTargets()
	if len(got) != len(want) {
		t.Fatalf("targets = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("targets = %#v, want %#v", got, want)
		}
	}
}

func TestBuildPHP83UsesEmbeddedRecipeAndExplicitImageTag(t *testing.T) {
	runner := &recordingRunner{}
	if err := BuildPHP83(context.Background(), runner); err != nil {
		t.Fatal(err)
	}
	args := runner.args[0]
	wantPrefix := []string{"image", "build", "--build-arg", "PHP_VERSION=8.3", "--tag", PHP83Image, "--file"}
	for i := range wantPrefix {
		if args[i] != wantPrefix[i] {
			t.Fatalf("args = %#v, want prefix %#v", args, wantPrefix)
		}
	}
	if len(args) != 9 || args[7] == "" || args[8] == "" {
		t.Fatalf("args = %#v; want recipe file and context directory", args)
	}
}

func TestPHPReadinessFailureIsActionable(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{ExitCode: 17}}
	err := (Podman{Runner: runner}).checkPHPReady(context.Background(), "chauf-php83-fpm")
	if err == nil || !strings.Contains(err.Error(), "php -v") || !strings.Contains(err.Error(), "17") {
		t.Fatalf("err = %v; want actionable PHP readiness failure", err)
	}
}

func TestInspectImageParsesDigestAndRecipeMetadata(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{Stdout: `[{"Id":"sha256:abc","RepoDigests":["ghcr.io/siegg/chauffeur-php@sha256:def"],"Architecture":"amd64","Labels":{"com.siegg.chauffeur.php.version":"8.3","com.siegg.chauffeur.php.base":"php:8.3-fpm-bookworm","com.siegg.chauffeur.php.extensions":"gd, pdo_pgsql, zip","com.siegg.chauffeur.recipe":"php83-v2"}}]`}}
	metadata, err := InspectImage(context.Background(), runner, PHP83Image)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Digest == "" || metadata.PHPVersion != "8.3" || metadata.RecipeFingerprint != "php83-v2" || len(metadata.Extensions) != 3 {
		t.Fatalf("metadata = %+v", metadata)
	}
	want := []string{"image", "inspect", "--format", "json", PHP83Image}
	for i := range want {
		if runner.args[0][i] != want[i] {
			t.Fatalf("args = %#v, want %#v", runner.args[0], want)
		}
	}
}

func TestEnsurePHPContainerStartsExistingContainer(t *testing.T) {
	runner := &existingPHPRunner{}
	r := Podman{Runner: runner}
	if err := r.EnsurePHPContainer(context.Background(), Scope{Version: "8.3"}, PHP83Image, "/tmp/project"); err != nil {
		t.Fatal(err)
	}
	if len(runner.args) != 9 {
		t.Fatalf("calls = %#v, want preflight, existence, inspect, and start checks", runner.args)
	}
	if runner.args[7][0] != "container" || runner.args[7][1] != "start" {
		t.Fatalf("container start args = %#v", runner.args[7])
	}
	if runner.args[8][0] != "container" || runner.args[8][1] != "exec" {
		t.Fatalf("PHP readiness args = %#v", runner.args[8])
	}
}

type existingPHPRunner struct{ args [][]string }

func (r *existingPHPRunner) Run(_ context.Context, args ...string) (CommandResult, error) {
	r.args = append(r.args, args)
	if len(args) >= 2 && args[0] == "container" && args[1] == "inspect" {
		return CommandResult{Stdout: `[{"State":{"Status":"exited"},"Config":{"Labels":{"com.siegg.chauffeur.role":"php-fpm","com.siegg.chauffeur.php.version":"8.3","com.siegg.chauffeur.scope":"shared"}},"Mounts":[{"Source":"/tmp/project","Destination":"/workspace","RW":true}]}]`}, nil
	}
	if len(args) >= 2 && args[0] == "image" && args[1] == "inspect" {
		return CommandResult{Stdout: `[{"Id":"sha256:test","Architecture":"amd64"}]`}, nil
	}
	return CommandResult{ExitCode: 0}, nil
}
