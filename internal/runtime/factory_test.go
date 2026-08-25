package runtime

import (
	"context"
	"testing"

	"github.com/siegg/chauffeur/internal/workspace"
)

func TestForWorkspaceDefaultsToNative(t *testing.T) {
	cfg := workspace.DefaultConfig("/tmp/chauffeur")
	r, err := ForWorkspace(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(Native); !ok {
		t.Fatalf("runtime = %T, want Native", r)
	}
}

func TestNativeAdapterSupportsProjectPreparation(t *testing.T) {
	cfg := workspace.DefaultConfig("/tmp/chauffeur")
	r, err := ForWorkspace(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.EnsureProject(context.Background(), Scope{Version: "8.3"}, PHP83Image, map[string]string{"/workspace": "/tmp/project"}); err != nil {
		t.Fatal(err)
	}
}

func TestForWorkspaceSelectsPodman(t *testing.T) {
	cfg := workspace.DefaultConfig("/tmp/chauffeur")
	cfg.Runtime.Engine = "podman"
	r, err := ForWorkspace(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.(Podman); !ok {
		t.Fatalf("runtime = %T, want Podman", r)
	}
}
