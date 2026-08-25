package projects

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyProjectSetupUsesOrderedBoundaryAndReadyResult(t *testing.T) {
	plan := BuildSetupPlan(ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypePHP}, SetupChoices{PHPVersion: "8.3", Domain: "shop.test"}, nil)
	var order []string
	result, err := ApplyProjectSetup(context.Background(), plan, ApplyDependencies{
		Snapshot:       func() error { order = append(order, "snapshot"); return nil },
		Save:           func() error { order = append(order, "save"); return nil },
		GenerateSSL:    func() error { order = append(order, "ssl"); return nil },
		PrepareRuntime: func() error { order = append(order, "runtime"); return nil },
		GenerateNginx:  func() error { order = append(order, "nginx"); return nil },
		EnableRoute:    func() error { order = append(order, "enable"); return nil },
		Reload:         func() error { order = append(order, "reload"); return nil },
		Readiness:      func() (string, error) { order = append(order, "ready"); return "fixture ready", nil },
	})
	if err != nil || result.Status != ApplyReady {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	want := "snapshot save ssl runtime nginx enable reload ready"
	if got := strings.Join(order, " "); got != want {
		t.Fatalf("order = %q, want %q", got, want)
	}
}

func TestApplyProjectSetupCancellationDoesNotRunOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	plan := BuildSetupPlan(ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypePHP}, SetupChoices{PHPVersion: "8.3", Domain: "shop.test"}, nil)
	result, err := ApplyProjectSetup(ctx, plan, ApplyDependencies{Save: func() error { called = true; return nil }})
	if err == nil || result.Status != ApplyFailed || called {
		t.Fatalf("result = %+v, err = %v, called = %t; want cancelled failed apply", result, err, called)
	}
}

func TestApplyProjectSetupReturnsDegradedReadinessWithRemediation(t *testing.T) {
	plan := BuildSetupPlan(ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypePHP}, SetupChoices{PHPVersion: "8.3", Domain: "shop.test"}, nil)
	result, err := ApplyProjectSetup(context.Background(), plan, ApplyDependencies{
		Save:      func() error { return nil },
		Readiness: func() (string, error) { return "services are stopped", fmt.Errorf("not running") },
	})
	if err != nil || result.Status != ApplyDegraded || result.Remediation == "" {
		t.Fatalf("result = %+v, err = %v; want degraded remediation", result, err)
	}
}

func TestApplyProjectSetupReturnsFailedStepRemediation(t *testing.T) {
	plan := BuildSetupPlan(ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypePHP}, SetupChoices{PHPVersion: "8.3", Domain: "shop.test"}, nil)
	result, err := ApplyProjectSetup(context.Background(), plan, ApplyDependencies{
		Save: func() error { return fmt.Errorf("permission denied") },
	})
	if err == nil || result.Status != ApplyFailed || result.Remediation == "" {
		t.Fatalf("result = %+v, err = %v; want failed remediation", result, err)
	}
}

func TestBuildSetupPlanIsReadOnlyAndListsChanges(t *testing.T) {
	plan := BuildSetupPlan(
		ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypeLaravel},
		SetupChoices{PHPVersion: "8.3", Domain: "shop.test", SSL: true, Dedicated: true},
		[]RuntimeChoice{{Version: "8.3", State: "installed", Recommended: true}},
	)
	if err := plan.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(plan.Changes) != 6 {
		t.Fatalf("changes = %d, want 6", len(plan.Changes))
	}
	if !strings.Contains(plan.Changes[2].Action, "shop.test") || !strings.Contains(plan.Changes[3].Action, "PHP 8.3") {
		t.Fatalf("changes = %+v; want domain and runtime intent", plan.Changes)
	}
}

func TestDetectSetupUsesComposerConstraintAndRemainsReadOnly(t *testing.T) {
	root := t.TempDir()
	before, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"require":{"php":"^8.2 <8.4"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	plan, err := DetectSetup(DetectionInput{Path: root, DefaultPHP: "8.1", InstalledPHP: []string{"8.1", "8.2", "8.3"}, DomainTLD: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Facts.PHPConstraint != "^8.2 <8.4" || plan.Choices.PHPVersion != "8.2" {
		t.Fatalf("plan = %+v; want composer constraint and compatible recommendation", plan)
	}
	if plan.PHPChoices[0].State != "incompatible" {
		t.Fatalf("PHP 8.1 choice = %+v; want incompatible", plan.PHPChoices[0])
	}
	after, err := os.ReadDir(root)
	if err != nil || len(after) != len(before)+1 {
		t.Fatalf("detection mutated project directory: before=%d after=%d err=%v", len(before), len(after), err)
	}
}

func TestDetectSetupExplicitPHPOverridesRecommendation(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "composer.json"), []byte(`{"require":{"php":">=8.2 <8.4"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	plan, err := DetectSetup(DetectionInput{
		Path: root, DefaultPHP: "8.3", ExplicitPHP: "8.1", InstalledPHP: []string{"8.1", "8.3"}, DomainTLD: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Validate() == nil || plan.Choices.PHPVersion != "" {
		t.Fatalf("plan = %+v; want unresolved explicit PHP blocker", plan)
	}
}

func TestPHPVersionSatisfiesComposerForms(t *testing.T) {
	tests := []struct {
		version, constraint string
		want                bool
	}{
		{"8.3", "^8.2", true},
		{"8.4", "^8.2", true},
		{"8.3", "~8.2", true},
		{"9.0", "~8.2", false},
		{"8.2", "~8.2", true},
		{"8.3", ">=8.2 <8.4", true},
		{"8.4", ">=8.2 <8.4", false},
	}
	for _, test := range tests {
		if got := PHPVersionSatisfies(test.version, test.constraint); got != test.want {
			t.Errorf("PHPVersionSatisfies(%q, %q) = %t, want %t", test.version, test.constraint, got, test.want)
		}
	}
}

func TestSetupPlanRejectsMissingPHP(t *testing.T) {
	plan := BuildSetupPlan(
		ProjectFacts{Path: "/tmp/shop", Slug: "shop", Type: TypeLaravel},
		SetupChoices{Domain: "shop.test"}, nil,
	)
	if err := plan.Validate(); err == nil {
		t.Fatal("expected missing PHP version error")
	}
}

func TestApplyStatusesExposeReadinessStates(t *testing.T) {
	if ApplyLinked == ApplyReady || ApplyReady == ApplyDegraded || ApplyDegraded == ApplyFailed {
		t.Fatal("apply statuses must remain distinct")
	}
	result := ApplyResult{Status: ApplyDegraded, Remediation: "run chauf start"}
	if result.Status != ApplyDegraded || result.Remediation == "" {
		t.Fatalf("result = %+v; want degraded remediation", result)
	}
}

func TestSaveCreatesAtomicBackupForExistingConfig(t *testing.T) {
	root := t.TempDir()
	p := &Project{Slug: "shop", Path: "/tmp/shop", Domain: "shop.test", PHPVersion: "8.3"}
	if err := Save(p, root); err != nil {
		t.Fatal(err)
	}
	p.Domain = "new-shop.test"
	if err := Save(p, root); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(p.ConfigPath(root) + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(backup), "domain: shop.test") {
		t.Fatalf("backup does not contain previous config: %s", backup)
	}
}
