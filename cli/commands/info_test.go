package commands

import (
	"path/filepath"
	"testing"
)

func TestRunInfoHelp(t *testing.T) {
	if err := RunInfo([]string{"--help"}); err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}
}

func TestShortSHA(t *testing.T) {
	if got := shortSHA("abcdef123456"); got != "abcdef1" {
		t.Fatalf("expected shortened sha, got %s", got)
	}
}

func TestDescribeBinaryServiceMissing(t *testing.T) {
	prefix := t.TempDir()
	svc := describeBinaryService("nginx", filepath.Join(prefix, "missing"), "-v")
	if svc.Status != "not installed" {
		t.Fatalf("expected not installed, got %s", svc.Status)
	}
}

func TestGatherBinaryServicesNoBinaries(t *testing.T) {
	prefix := t.TempDir()
	services := gatherBinaryServices(prefix)
	if len(services) != 2 {
		t.Fatalf("expected two services (nginx/composer) entries, got %d", len(services))
	}
	for _, svc := range services {
		if svc.Status != "not installed" {
			t.Fatalf("expected not installed, got %s", svc.Status)
		}
	}
}

func TestSettersForBuildMetadata(t *testing.T) {
	SetCLIVersion("1.0.0")
	SetBuildCommit("deadbeef")
	SetBuildTimestamp("now")

	if getCLIVersion() != "1.0.0" || getBuildCommit() != "deadbeef" || getBuildTimestamp() != "now" {
		t.Fatalf("metadata setters failed")
	}
}
