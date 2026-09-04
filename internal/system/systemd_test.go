package system

import (
	"strings"
	"testing"
)

func TestPodmanUnitsUseContainerLifecycleAndDependencyOrder(t *testing.T) {
	fpm := PodmanFPMUnit("chauf-php83-fpm")
	nginx := PodmanNginxUnitContent([]string{fpm})
	if !strings.Contains(nginx, "Requires="+fpm) || !strings.Contains(nginx, "After="+fpm) {
		t.Fatalf("nginx unit lacks FPM dependency: %s", nginx)
	}
	if strings.Contains(nginx, ".chauffeur/nginx") || strings.Contains(nginx, "podman start chauf-nginx") {
		t.Fatalf("Podman nginx unit contains native paths: %s", nginx)
	}
	fpmContent := PodmanFPMUnitContent("chauf-php83-fpm")
	if !strings.Contains(fpmContent, "podman start chauf-php83-fpm") {
		t.Fatalf("FPM unit does not start the Podman container: %s", fpmContent)
	}
}

func TestNativeUnitsUseResolvedWorkspaceRoot(t *testing.T) {
	root := "/tmp/custom-chauffeur"
	nginx := NginxUnitContentForRoot(root)
	if !strings.Contains(nginx, "ExecStart="+root+"/nginx/sbin/nginx") || strings.Contains(nginx, "%h/.chauffeur") {
		t.Fatalf("nginx unit does not use custom root: %s", nginx)
	}
	fpm := FPMTemplateUnitContentForRoot(root)
	if !strings.Contains(fpm, "Before=chauffeur-nginx.service") || !strings.Contains(fpm, root+"/php/%i/sbin/php-fpm") {
		t.Fatalf("FPM unit has incorrect root/order: %s", fpm)
	}
}
