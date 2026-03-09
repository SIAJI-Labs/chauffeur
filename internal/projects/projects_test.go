package projects_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/siegg/chauffeur/internal/projects"
)

// ── GenerateSlug ──────────────────────────────────────────────────────────────

func TestGenerateSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/home/user/Projects/my-app", "my-app"},
		{"/home/user/Projects/My App", "my-app"},
		{"/home/user/Projects/my_project", "my-project"},
		{"/home/user/Projects/My_App_V2", "my-app-v2"},
		{"/home/user/Projects/hello--world", "hello-world"},
		{"/home/user/Projects/-leading", "leading"},
		{"/home/user/Projects/trailing-", "trailing"},
		{"/home/user/Projects/UPPERCASE", "uppercase"},
		{"/home/user/Projects/dots.and.dots", "dotsanddots"},
	}
	for _, tc := range cases {
		got := projects.GenerateSlug(tc.in)
		if got != tc.want {
			t.Errorf("GenerateSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestGenerateSlug_MaxLength(t *testing.T) {
	long := "/home/user/" + strings.Repeat("a", 100)
	got := projects.GenerateSlug(long)
	if len(got) > 50 {
		t.Errorf("slug length %d exceeds max 50", len(got))
	}
}

// ── IsValidDomain ─────────────────────────────────────────────────────────────

func TestIsValidDomain(t *testing.T) {
	valid := []string{
		"my-project.test",
		"admin.my-project.test",
		"api.v2.my-project.test",
		"a.test",
	}
	for _, d := range valid {
		if !projects.IsValidDomain(d) {
			t.Errorf("IsValidDomain(%q) = false, want true", d)
		}
	}

	invalid := []string{
		"my-project.local",   // wrong TLD
		"my-project.com",
		"-bad.test",          // leading hyphen
		"bad-.test",          // trailing hyphen
		"My-Project.test",    // uppercase
		".test",              // no label before tld
		"my project.test",    // space
		"",
	}
	for _, d := range invalid {
		if projects.IsValidDomain(d) {
			t.Errorf("IsValidDomain(%q) = true, want false", d)
		}
	}
}

// ── DomainFromSlug ────────────────────────────────────────────────────────────

func TestDomainFromSlug(t *testing.T) {
	if got := projects.DomainFromSlug("my-app", "test"); got != "my-app.test" {
		t.Errorf("got %q, want %q", got, "my-app.test")
	}
}

// ── Detect ────────────────────────────────────────────────────────────────────

func TestDetect_Laravel(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "artisan"), []byte(""), 0644)
	if got := projects.Detect(dir); got != projects.TypeLaravel {
		t.Errorf("Detect = %q, want %q", got, projects.TypeLaravel)
	}
}

func TestDetect_WordPress(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "wp-config.php"), []byte(""), 0644)
	if got := projects.Detect(dir); got != projects.TypeWordPress {
		t.Errorf("Detect = %q, want %q", got, projects.TypeWordPress)
	}
}

func TestDetect_Generic(t *testing.T) {
	dir := t.TempDir()
	if got := projects.Detect(dir); got != projects.TypeGeneric {
		t.Errorf("Detect = %q, want %q", got, projects.TypeGeneric)
	}
}

// ── DocumentRoot ──────────────────────────────────────────────────────────────

func TestDocumentRoot_Laravel(t *testing.T) {
	dir := t.TempDir()
	got := projects.DocumentRoot(dir, projects.TypeLaravel)
	if got != filepath.Join(dir, "public") {
		t.Errorf("DocumentRoot Laravel = %q, want %q", got, filepath.Join(dir, "public"))
	}
}

func TestDocumentRoot_WordPress(t *testing.T) {
	dir := t.TempDir()
	got := projects.DocumentRoot(dir, projects.TypeWordPress)
	if got != dir {
		t.Errorf("DocumentRoot WordPress = %q, want %q", got, dir)
	}
}

func TestDocumentRoot_Generic_WithPublic(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "public"), 0755)
	got := projects.DocumentRoot(dir, projects.TypeGeneric)
	if got != filepath.Join(dir, "public") {
		t.Errorf("DocumentRoot Generic+public = %q, want %q", got, filepath.Join(dir, "public"))
	}
}

func TestDocumentRoot_Generic_NoPublic(t *testing.T) {
	dir := t.TempDir()
	got := projects.DocumentRoot(dir, projects.TypeGeneric)
	if got != dir {
		t.Errorf("DocumentRoot Generic (no public) = %q, want %q", got, dir)
	}
}

// ── Save / Load round-trip ────────────────────────────────────────────────────

func newTestProject() *projects.Project {
	return &projects.Project{
		Slug:        "my-app",
		Path:        "/home/user/Projects/my-app",
		Domain:      "my-app.test",
		Aliases:     []string{"admin.my-app.test", "api.my-app.test"},
		PHPVersion:  "8.3",
		SSL:         false,
		FPM:         projects.FPMConfig{Dedicated: false, Socket: ""},
		ProjectType: projects.TypeLaravel,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()

	if err := projects.Save(p, root); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := projects.Load(root, p.Slug)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.Slug != p.Slug {
		t.Errorf("Slug: got %q, want %q", got.Slug, p.Slug)
	}
	if got.Path != p.Path {
		t.Errorf("Path: got %q, want %q", got.Path, p.Path)
	}
	if got.Domain != p.Domain {
		t.Errorf("Domain: got %q, want %q", got.Domain, p.Domain)
	}
	if len(got.Aliases) != 2 || got.Aliases[0] != "admin.my-app.test" {
		t.Errorf("Aliases: got %v, want %v", got.Aliases, p.Aliases)
	}
	if got.PHPVersion != "8.3" {
		t.Errorf("PHPVersion: got %q, want %q", got.PHPVersion, "8.3")
	}
	if got.SSL != false {
		t.Errorf("SSL: got %v, want false", got.SSL)
	}
	if got.FPM.Dedicated != false {
		t.Errorf("FPM.Dedicated: got %v, want false", got.FPM.Dedicated)
	}
	if got.ProjectType != projects.TypeLaravel {
		t.Errorf("ProjectType: got %q, want %q", got.ProjectType, projects.TypeLaravel)
	}
}

func TestSaveLoad_SSL(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	p.SSL = true
	p.FPM.Dedicated = true
	p.FPM.Socket = "/tmp/test.sock"

	if err := projects.Save(p, root); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := projects.Load(root, p.Slug)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !got.SSL {
		t.Error("SSL: got false, want true")
	}
	if !got.FPM.Dedicated {
		t.Error("FPM.Dedicated: got false, want true")
	}
	if got.FPM.Socket != "/tmp/test.sock" {
		t.Errorf("FPM.Socket: got %q, want %q", got.FPM.Socket, "/tmp/test.sock")
	}
}

func TestSaveLoad_EmptyAliases(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	p.Aliases = nil

	if err := projects.Save(p, root); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := projects.Load(root, p.Slug)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Aliases) != 0 {
		t.Errorf("Aliases: got %v, want nil/empty", got.Aliases)
	}
}

// ── ListSlugs / ListAll ───────────────────────────────────────────────────────

func TestListSlugs(t *testing.T) {
	root := t.TempDir()

	p1 := newTestProject()
	p1.Slug = "alpha"
	p1.Domain = "alpha.test"

	p2 := newTestProject()
	p2.Slug = "beta"
	p2.Domain = "beta.test"

	projects.Save(p1, root)
	projects.Save(p2, root)

	slugs, err := projects.ListSlugs(root)
	if err != nil {
		t.Fatalf("ListSlugs: %v", err)
	}
	if len(slugs) != 2 {
		t.Errorf("len(slugs) = %d, want 2", len(slugs))
	}
}

func TestListSlugs_Empty(t *testing.T) {
	root := t.TempDir()
	slugs, err := projects.ListSlugs(root)
	if err != nil {
		t.Fatalf("ListSlugs on empty workspace: %v", err)
	}
	if len(slugs) != 0 {
		t.Errorf("expected 0 slugs, got %d", len(slugs))
	}
}

// ── FindByPath ────────────────────────────────────────────────────────────────

func TestFindByPath_Found(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	got, err := projects.FindByPath(root, p.Path)
	if err != nil {
		t.Fatalf("FindByPath: %v", err)
	}
	if got == nil || got.Slug != p.Slug {
		t.Errorf("FindByPath: got %v, want slug %q", got, p.Slug)
	}
}

func TestFindByPath_NotFound(t *testing.T) {
	root := t.TempDir()
	got, err := projects.FindByPath(root, "/nonexistent/path")
	if err != nil {
		t.Fatalf("FindByPath: %v", err)
	}
	if got != nil {
		t.Errorf("FindByPath: expected nil, got %v", got)
	}
}

// ── IsDomainInUse ─────────────────────────────────────────────────────────────

func TestIsDomainInUse_Primary(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	owner, err := projects.IsDomainInUse(root, "my-app.test", "")
	if err != nil {
		t.Fatalf("IsDomainInUse: %v", err)
	}
	if owner != "my-app" {
		t.Errorf("owner = %q, want %q", owner, "my-app")
	}
}

func TestIsDomainInUse_Alias(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	owner, err := projects.IsDomainInUse(root, "admin.my-app.test", "")
	if err != nil {
		t.Fatalf("IsDomainInUse: %v", err)
	}
	if owner != "my-app" {
		t.Errorf("owner = %q, want %q", owner, "my-app")
	}
}

func TestIsDomainInUse_Free(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	owner, err := projects.IsDomainInUse(root, "other.test", "")
	if err != nil {
		t.Fatalf("IsDomainInUse: %v", err)
	}
	if owner != "" {
		t.Errorf("expected no owner, got %q", owner)
	}
}

func TestIsDomainInUse_SkipSelf(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	// Should not conflict with itself when updating
	owner, err := projects.IsDomainInUse(root, "my-app.test", "my-app")
	if err != nil {
		t.Fatalf("IsDomainInUse: %v", err)
	}
	if owner != "" {
		t.Errorf("expected no owner (skipSlug=my-app), got %q", owner)
	}
}

// ── AllDomains ────────────────────────────────────────────────────────────────

func TestAllDomains(t *testing.T) {
	p := &projects.Project{
		Domain:  "my-app.test",
		Aliases: []string{"admin.my-app.test", "api.my-app.test"},
	}
	got := p.AllDomains()
	if len(got) != 3 || got[0] != "my-app.test" {
		t.Errorf("AllDomains = %v, want [my-app.test admin.my-app.test api.my-app.test]", got)
	}
}

func TestAllDomains_NoAliases(t *testing.T) {
	p := &projects.Project{Domain: "solo.test"}
	got := p.AllDomains()
	if len(got) != 1 || got[0] != "solo.test" {
		t.Errorf("AllDomains = %v, want [solo.test]", got)
	}
}

// ── FPMSocketPath ─────────────────────────────────────────────────────────────

func TestFPMSocketPath_Shared(t *testing.T) {
	root := "/home/user/.chauffeur"
	p := &projects.Project{PHPVersion: "8.3", FPM: projects.FPMConfig{}}
	got := p.FPMSocketPath(root)
	want := root + "/php/8.3/runtime/php-fpm/php-fpm.sock"
	if got != want {
		t.Errorf("FPMSocketPath = %q, want %q", got, want)
	}
}

func TestFPMSocketPath_Dedicated(t *testing.T) {
	root := "/home/user/.chauffeur"
	p := &projects.Project{
		PHPVersion: "8.3",
		FPM:        projects.FPMConfig{Dedicated: true, Socket: root + "/projects/my-app/php-fpm.sock"},
	}
	got := p.FPMSocketPath(root)
	want := root + "/projects/my-app/php-fpm.sock"
	if got != want {
		t.Errorf("FPMSocketPath = %q, want %q", got, want)
	}
}

// ── RenderNginxConfig ─────────────────────────────────────────────────────────

func TestRenderNginxConfig_HTTP_Laravel(t *testing.T) {
	root := t.TempDir()
	p := &projects.Project{
		Slug:        "my-app",
		Path:        "/home/user/Projects/my-app",
		Domain:      "my-app.test",
		Aliases:     []string{"admin.my-app.test"},
		PHPVersion:  "8.3",
		SSL:         false,
		ProjectType: projects.TypeLaravel,
	}

	out, err := projects.RenderNginxConfig(p, root, 8080, 8443)
	if err != nil {
		t.Fatalf("RenderNginxConfig: %v", err)
	}

	checks := []string{
		"listen 8080",
		"server_name my-app.test admin.my-app.test",
		"root /home/user/Projects/my-app/public",
		"fastcgi_pass unix:" + root + "/php/8.3/runtime/php-fpm/php-fpm.sock",
		"$query_string",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("config missing %q\n\nGot:\n%s", want, out)
		}
	}
	if strings.Contains(out, "ssl_certificate") {
		t.Error("HTTP config should not contain ssl_certificate")
	}
}

func TestRenderNginxConfig_HTTP_WordPress(t *testing.T) {
	root := t.TempDir()
	p := &projects.Project{
		Slug:        "wp-site",
		Path:        "/home/user/Projects/wp-site",
		Domain:      "wp-site.test",
		PHPVersion:  "8.1",
		SSL:         false,
		ProjectType: projects.TypeWordPress,
	}

	out, err := projects.RenderNginxConfig(p, root, 8080, 8443)
	if err != nil {
		t.Fatalf("RenderNginxConfig: %v", err)
	}

	if !strings.Contains(out, "$args") {
		t.Error("WordPress config should use $args, not $query_string")
	}
	// WordPress root is the project dir, not public/
	if !strings.Contains(out, "root /home/user/Projects/wp-site;") {
		t.Errorf("WordPress document root should be project root\n\nGot:\n%s", out)
	}
}

func TestRenderNginxConfig_HTTPS(t *testing.T) {
	root := t.TempDir()
	p := &projects.Project{
		Slug:        "secure-app",
		Path:        "/home/user/Projects/secure-app",
		Domain:      "secure-app.test",
		PHPVersion:  "8.3",
		SSL:         true,
		ProjectType: projects.TypeLaravel,
	}

	out, err := projects.RenderNginxConfig(p, root, 8080, 8443)
	if err != nil {
		t.Fatalf("RenderNginxConfig: %v", err)
	}

	checks := []string{
		"listen 8080",
		"return 301 https://$host:8443$request_uri",
		"listen 8443 ssl",
		"http2 on",
		"ssl_certificate",
		"ssl_certificate_key",
		"ssl_protocols TLSv1.2 TLSv1.3",
		"fastcgi_param HTTPS on",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("HTTPS config missing %q\n\nGot:\n%s", want, out)
		}
	}
}

func TestWriteNginxConfig(t *testing.T) {
	root := t.TempDir()
	// Pre-create directories that WriteNginxConfig expects to create
	p := &projects.Project{
		Slug:        "test-proj",
		Path:        "/home/user/Projects/test-proj",
		Domain:      "test-proj.test",
		PHPVersion:  "8.3",
		ProjectType: projects.TypeGeneric,
	}

	if err := projects.WriteNginxConfig(p, root, 8080, 8443); err != nil {
		t.Fatalf("WriteNginxConfig: %v", err)
	}

	confPath := filepath.Join(root, "nginx", "etc", "sites-available", "test-proj.conf")
	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if !strings.Contains(string(data), "server_name test-proj.test") {
		t.Errorf("config file missing server_name\n\nGot:\n%s", data)
	}
}

func TestEnableDisableNginxSite(t *testing.T) {
	root := t.TempDir()
	p := &projects.Project{
		Slug:        "link-test",
		Path:        "/tmp/link-test",
		Domain:      "link-test.test",
		PHPVersion:  "8.3",
		ProjectType: projects.TypeGeneric,
	}

	// Write the config first so the symlink target exists
	if err := projects.WriteNginxConfig(p, root, 8080, 8443); err != nil {
		t.Fatalf("WriteNginxConfig: %v", err)
	}

	if err := projects.EnableNginxSite(p, root); err != nil {
		t.Fatalf("EnableNginxSite: %v", err)
	}

	link := filepath.Join(root, "nginx", "etc", "sites-enabled", "link-test.conf")
	if _, err := os.Lstat(link); err != nil {
		t.Errorf("symlink not created: %v", err)
	}

	// Idempotent: enable again should not error
	if err := projects.EnableNginxSite(p, root); err != nil {
		t.Errorf("EnableNginxSite (idempotent): %v", err)
	}

	if err := projects.DisableNginxSite(p, root); err != nil {
		t.Fatalf("DisableNginxSite: %v", err)
	}
	if _, err := os.Lstat(link); err == nil {
		t.Error("symlink should be removed after DisableNginxSite")
	}

	// Disable again on missing link should not error
	if err := projects.DisableNginxSite(p, root); err != nil {
		t.Errorf("DisableNginxSite (already removed): %v", err)
	}
}

func TestDelete(t *testing.T) {
	root := t.TempDir()
	p := newTestProject()
	projects.Save(p, root)

	if err := projects.Delete(root, p.Slug); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := projects.Load(root, p.Slug); err == nil {
		t.Error("expected error loading deleted project, got nil")
	}
}
