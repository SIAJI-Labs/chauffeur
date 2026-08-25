package runtime

import "testing"

func TestPHPImageUsesNormalizedFriendlyVersion(t *testing.T) {
	if got := PHPImage(" 8.3 "); got != "ghcr.io/siegg/chauffeur-php:8.3-fpm" {
		t.Fatalf("image = %q", got)
	}
}
