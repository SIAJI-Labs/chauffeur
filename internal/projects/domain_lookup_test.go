package projects

import "testing"

func TestFindByDomainMatchesPrimaryDomainAndAliasForMissingPath(t *testing.T) {
	root := t.TempDir()
	project := &Project{Slug: "stale", Path: "/deleted/project", Domain: "stale.test", Aliases: []string{"www.stale.test"}, PHPVersion: "8.3", ProjectType: TypePHP}
	if err := Save(project, root); err != nil {
		t.Fatal(err)
	}
	for _, domain := range []string{"stale.test", "WWW.STALE.TEST"} {
		got, err := FindByDomain(root, domain)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil || got.Slug != project.Slug {
			t.Fatalf("FindByDomain(%q) = %+v; want stale project", domain, got)
		}
	}
}
