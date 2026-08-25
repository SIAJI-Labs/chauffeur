package services

import (
	"testing"

	"github.com/siegg/chauffeur/internal/projects"
)

func TestAllFPMSkipsReverseProxyProjects(t *testing.T) {
	root := t.TempDir()
	for _, project := range []*projects.Project{
		{Slug: "proxy", Path: "/tmp/proxy", ProjectType: projects.TypeReverseProxy},
		{Slug: "php", Path: "/tmp/php", PHPVersion: "8.3", ProjectType: projects.TypePHP},
	} {
		if err := projects.Save(project, root); err != nil {
			t.Fatalf("save project %s: %v", project.Slug, err)
		}
	}

	fpms, err := NewManager(root).AllFPM()
	if err != nil {
		t.Fatalf("AllFPM: %v", err)
	}
	if len(fpms) != 1 {
		t.Fatalf("AllFPM returned %d pools; want 1", len(fpms))
	}
	if got := fpms[0].Label(); got != "8.3 (shared)" {
		t.Fatalf("pool label = %q; want 8.3 (shared)", got)
	}
}
