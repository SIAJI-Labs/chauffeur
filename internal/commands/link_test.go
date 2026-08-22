package commands

import (
	"strings"
	"testing"

	"github.com/siegg/chauffeur/internal/projects"
	"github.com/siegg/chauffeur/internal/workspace"
)

func summaryMap(items []unlinkSummaryItem) map[string]string {
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.label] = item.value
	}
	return values
}

func TestUnlinkSummary_ReverseProxy(t *testing.T) {
	cfg := workspace.DefaultConfig(t.TempDir())
	p := &projects.Project{
		Path:        "/home/user/Projects/arvi-ui",
		Domain:      "arvi-ui.test",
		Aliases:     []string{"app.arvi-ui.test"},
		ProjectType: projects.TypeReverseProxy,
		ProxyPort:   3901,
	}

	values := summaryMap(unlinkSummary(p, cfg))
	checks := map[string]string{
		"Domain":       "http://arvi-ui.test:8080",
		"SSL":          "Disabled (HTTP only)",
		"Type":         "reverse proxy",
		"Proxy target": "http://localhost:3901",
		"Runtime":      "No PHP-FPM; external development server",
	}
	for label, want := range checks {
		if values[label] != want {
			t.Errorf("%s = %q, want %q", label, values[label], want)
		}
	}
	if !strings.Contains(values["Nginx"], "removed") {
		t.Errorf("Nginx summary = %q, want removal warning", values["Nginx"])
	}
}

func TestUnlinkSummary_SSLPHPProject(t *testing.T) {
	cfg := workspace.DefaultConfig(t.TempDir())
	p := &projects.Project{
		Path:        "/home/user/Projects/shop",
		Domain:      "shop.test",
		SSL:         true,
		PHPVersion:  "8.3",
		ProjectType: projects.TypeLaravel,
		FPM:         projects.FPMConfig{Dedicated: true},
	}

	values := summaryMap(unlinkSummary(p, cfg))
	if values["Domain"] != "https://shop.test:8443" {
		t.Errorf("Domain = %q, want HTTPS URL", values["Domain"])
	}
	if !strings.Contains(values["SSL"], "Enabled") ||
		!strings.Contains(values["SSL"], "certificate retained") {
		t.Errorf("SSL summary = %q, want enabled and retained certificate", values["SSL"])
	}
	if values["Runtime"] != "PHP 8.3 · dedicated FPM" {
		t.Errorf("Runtime = %q, want PHP/FPM details", values["Runtime"])
	}
}

func TestProjectDetailRuntime_ReverseProxyDoesNotUsePHPFPM(t *testing.T) {
	p := &projects.Project{
		Path:        "/home/user/Projects/arvi-ui",
		ProjectType: projects.TypeReverseProxy,
		ProxyPort:   3901,
	}
	label, value := projectDetailRuntime(p)
	if label != "Proxy" || value != "http://localhost:3901" {
		t.Errorf("runtime = %s %q, want Proxy http://localhost:3901", label, value)
	}
}

func TestProjectDetailRuntime_PHPUsesFPM(t *testing.T) {
	p := &projects.Project{
		ProjectType: projects.TypeLaravel,
		PHPVersion:  "8.3",
		FPM:         projects.FPMConfig{Dedicated: true},
	}
	label, value := projectDetailRuntime(p)
	if label != "PHP" || value != "8.3 (dedicated FPM)" {
		t.Errorf("runtime = %s %q, want PHP 8.3 (dedicated FPM)", label, value)
	}
}
