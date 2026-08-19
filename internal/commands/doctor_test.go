package commands

import (
	"strings"
	"testing"
)

func TestDoctorPHPDepsIncludesPostgresClientLibrary(t *testing.T) {
	results := doctorPHPDeps()
	var found bool
	for _, r := range results {
		if r.name == "libpq" {
			found = true
			if r.warn {
				t.Fatalf("libpq check should be a hard failure, got warn=true: %#v", r)
			}
			if !strings.Contains(r.fix, "postgresql-libs") && !strings.Contains(r.fix, "libpq-dev") && !strings.Contains(r.fix, "postgresql-devel") {
				t.Fatalf("libpq fix should mention a PostgreSQL client package, got %q", r.fix)
			}
		}
	}
	if !found {
		t.Fatal("expected doctorPHPDeps to include libpq")
	}
}
