package commands

import (
	"testing"

	"github.com/siegg/chauffeur/internal/projects"
)

func TestEnsureSSLDegradesWhenMkcertIsUnavailable(t *testing.T) {
	if projects.MkcertInstalled() {
		t.Skip("mkcert is available in this environment")
	}
	p := &projects.Project{Domain: "shop.test", SSL: true}
	if err := ensureSSL(p, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if p.SSL {
		t.Fatal("SSL should be disabled when mkcert is unavailable")
	}
}
