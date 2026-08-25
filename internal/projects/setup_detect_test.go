package projects

import "testing"

func TestDetectSetupIsReadOnlyAndUsesExistingIntent(t *testing.T) {
	plan, err := DetectSetup(DetectionInput{
		Path:         t.TempDir(),
		DefaultPHP:   "8.3",
		InstalledPHP: []string{"8.3", "8.2"},
		DomainTLD:    "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Facts.Slug == "" || plan.Facts.DocumentRoot == "" || len(plan.Facts.Evidence) == 0 || len(plan.PHPChoices) != 2 {
		t.Fatalf("unexpected detection plan: %+v", plan)
	}
	if !plan.PHPChoices[0].Recommended {
		t.Fatal("default PHP should be recommended")
	}
}

func TestDetectSetupPreservesExistingRelinkIntent(t *testing.T) {
	root := t.TempDir()
	existing := &Project{
		Slug: "shop", Path: root, Domain: "shop.test", Aliases: []string{"admin.shop.test"},
		PHPVersion: "8.2", SSL: true, FPM: FPMConfig{Dedicated: true}, ProjectType: TypePHP,
	}
	plan, err := DetectSetup(DetectionInput{Path: root, Existing: existing, DefaultPHP: "8.3", InstalledPHP: []string{"8.2", "8.3"}, DomainTLD: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Choices.Domain != existing.Domain || len(plan.Choices.Aliases) != 1 || plan.Choices.PHPVersion != existing.PHPVersion || !plan.Choices.SSL || !plan.Choices.Dedicated {
		t.Fatalf("relink choices = %+v; want existing intent preserved", plan.Choices)
	}
}

func TestDetectSetupPrefersExistingPHPBeforeOtherCompatibleRuntime(t *testing.T) {
	root := t.TempDir()
	plan, err := DetectSetup(DetectionInput{
		Path: root, DefaultPHP: "8.3", InstalledPHP: []string{"8.2", "8.3"}, DomainTLD: "test",
		Existing: &Project{Slug: "shop", Path: root, Domain: "shop.test", PHPVersion: "8.3", ProjectType: TypePHP},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Choices.PHPVersion != "8.3" {
		t.Fatalf("recommended PHP = %q; want existing 8.3", plan.Choices.PHPVersion)
	}
}

func TestDetectSetupWarnsWhenSSLReadinessIsMissing(t *testing.T) {
	root := t.TempDir()
	plan, err := DetectSetup(DetectionInput{Path: root, DefaultPHP: "8.3", InstalledPHP: []string{"8.3"}, DomainTLD: "test", Existing: &Project{Slug: "shop", Path: root, Domain: "shop.test", PHPVersion: "8.3", SSL: true, ProjectType: TypePHP}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Warnings) != 1 || plan.Warnings[0].Code != "ssl-not-ready" {
		t.Fatalf("warnings = %+v; want SSL readiness warning", plan.Warnings)
	}
}
