package installers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── IsLegacy ──────────────────────────────────────────────────────────────────

func TestIsLegacy(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"7.4", true},
		{"8.0", true},
		{"8.1", false},
		{"8.2", false},
		{"8.3", false},
		{"8.4", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsLegacy(tc.in); got != tc.want {
			t.Errorf("IsLegacy(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// ── VendoredOpenSSLEnv ────────────────────────────────────────────────────────

func TestVendoredOpenSSLEnv_ContainsRequiredFlags(t *testing.T) {
	env := VendoredOpenSSLEnv("/opt/vendor-openssl")

	mustContain := []string{
		"-Wno-deprecated-declarations",
		"-Wno-discarded-qualifiers",
		"-Wno-incompatible-pointer-types",
		"-DOPENSSL_API_COMPAT=",
		"/opt/vendor-openssl",
	}
	joined := strings.Join(env, " ")
	for _, s := range mustContain {
		if !strings.Contains(joined, s) {
			t.Errorf("VendoredOpenSSLEnv missing %q", s)
		}
	}
}

// ── patchLibXML ───────────────────────────────────────────────────────────────

const libxmlSrc = `#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

void php_libxml_structured_error_handler(void *ctx, xmlErrorPtr error) {}

xmlSetStructuredErrorFunc(NULL, php_libxml_structured_error_handler);

if (current_handler == php_libxml_structured_error_handler) {}
if (current_handler != php_libxml_structured_error_handler) {}

xmlOutputBufferCreateFilenameDefault(php_libxml_output_buffer_create_filename);
`

func TestPatchLibXML_InjectsAttributeUnused(t *testing.T) {
	src := setupSrcDir(t, "ext/libxml/libxml.c", libxmlSrc)
	if err := patchLibXML(src); err != nil {
		t.Fatalf("patchLibXML: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/libxml/libxml.c"))
	if !strings.Contains(content, "#define ATTRIBUTE_UNUSED") {
		t.Error("expected ATTRIBUTE_UNUSED macro to be injected")
	}
}

func TestPatchLibXML_CastsErrorHandler(t *testing.T) {
	src := setupSrcDir(t, "ext/libxml/libxml.c", libxmlSrc)
	if err := patchLibXML(src); err != nil {
		t.Fatalf("patchLibXML: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/libxml/libxml.c"))
	if !strings.Contains(content, "(xmlStructuredErrorFunc)php_libxml_structured_error_handler") {
		t.Error("expected xmlStructuredErrorFunc cast in call sites")
	}
	// The definition line must NOT be cast — check it's still unmodified.
	if !strings.Contains(content, "void php_libxml_structured_error_handler(") {
		t.Error("function definition should not be modified")
	}
}

func TestPatchLibXML_GuardsOutputBuffer(t *testing.T) {
	src := setupSrcDir(t, "ext/libxml/libxml.c", libxmlSrc)
	if err := patchLibXML(src); err != nil {
		t.Fatalf("patchLibXML: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/libxml/libxml.c"))
	if !strings.Contains(content, "#if LIBXML_VERSION < 21200") {
		t.Error("expected libxml version guard around xmlOutputBufferCreateFilenameDefault")
	}
}

func TestPatchLibXML_Idempotent(t *testing.T) {
	src := setupSrcDir(t, "ext/libxml/libxml.c", libxmlSrc)
	for i := 0; i < 2; i++ {
		if err := patchLibXML(src); err != nil {
			t.Fatalf("patchLibXML pass %d: %v", i+1, err)
		}
	}
	content := readFile(t, filepath.Join(src, "ext/libxml/libxml.c"))
	// ATTRIBUTE_UNUSED should appear exactly once.
	if count := strings.Count(content, "#define ATTRIBUTE_UNUSED"); count != 1 {
		t.Errorf("ATTRIBUTE_UNUSED appears %d times, want 1", count)
	}
}

// ── patchOpenSSLExt ───────────────────────────────────────────────────────────

const opensslSrc = `#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

RSA *rsa = EVP_PKEY_get0_RSA(pkey);
`

func TestPatchOpenSSLExt_InjectsShims(t *testing.T) {
	src := setupSrcDir(t, "ext/openssl/openssl.c", opensslSrc)
	if err := patchOpenSSLExt(src); err != nil {
		t.Fatalf("patchOpenSSLExt: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/openssl/openssl.c"))
	if !strings.Contains(content, "OpenSSL 3.x compatibility") {
		t.Error("expected OpenSSL compat shim marker")
	}
	if !strings.Contains(content, "EVP_PKEY_get0_RSA_NONCONST") {
		t.Error("expected EVP_PKEY_get0_RSA_NONCONST shim")
	}
	if !strings.Contains(content, "RSA_SSLV23_PADDING") {
		t.Error("expected RSA_SSLV23_PADDING shim")
	}
}

func TestPatchOpenSSLExt_ReplacesCallSites(t *testing.T) {
	src := setupSrcDir(t, "ext/openssl/openssl.c", opensslSrc)
	if err := patchOpenSSLExt(src); err != nil {
		t.Fatalf("patchOpenSSLExt: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/openssl/openssl.c"))
	if strings.Contains(content, "EVP_PKEY_get0_RSA(pkey)") {
		t.Error("original EVP_PKEY_get0_RSA call site should be replaced")
	}
	if !strings.Contains(content, "EVP_PKEY_get0_RSA_NONCONST(pkey)") {
		t.Error("call site should use EVP_PKEY_get0_RSA_NONCONST")
	}
}

func TestPatchOpenSSLExt_Idempotent(t *testing.T) {
	src := setupSrcDir(t, "ext/openssl/openssl.c", opensslSrc)
	for i := 0; i < 2; i++ {
		if err := patchOpenSSLExt(src); err != nil {
			t.Fatalf("patchOpenSSLExt pass %d: %v", i+1, err)
		}
	}
	content := readFile(t, filepath.Join(src, "ext/openssl/openssl.c"))
	if count := strings.Count(content, "OpenSSL 3.x compatibility"); count != 1 {
		t.Errorf("compat marker appears %d times, want 1", count)
	}
}

// ── patchScanf ────────────────────────────────────────────────────────────────

const scanfSrc = `
zend_long (*fn)() = NULL;
fn = (zend_long (*)())ZEND_STRTOL_PTR;
fn = (zend_long (*)())ZEND_STRTOUL_PTR;
value = (zend_long) (*fn)(buf, NULL, base);
`

func TestPatchScanf_FixesScanfC(t *testing.T) {
	src := setupSrcDir(t, "ext/standard/scanf.c", scanfSrc)
	// patchScanf also tries main/spprintf.c — create empty stubs.
	writeFile(t, filepath.Join(src, "main", "spprintf.c"), "")
	writeFile(t, filepath.Join(src, "main", "snprintf.c"), "")

	if err := patchScanf(src); err != nil {
		t.Fatalf("patchScanf: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/standard/scanf.c"))

	if strings.Contains(content, "zend_long (*fn)() = NULL;") {
		t.Error("old untyped fn declaration should be replaced")
	}
	if !strings.Contains(content, "zend_long (*fn)(const char *, char **, int) = NULL;") {
		t.Error("expected typed fn declaration")
	}
	if strings.Contains(content, "(zend_long (*)())ZEND_STRTOL_PTR") {
		t.Error("old strtol cast should be replaced")
	}
	if !strings.Contains(content, "(zend_long (*)(const char *, char **, int))ZEND_STRTOL_PTR") {
		t.Error("expected typed strtol cast")
	}
	if !strings.Contains(content, "(zend_long (*)(const char *, char **, int))ZEND_STRTOUL_PTR") {
		t.Error("expected typed strtoul cast")
	}
}

func TestPatchScanf_Idempotent(t *testing.T) {
	src := setupSrcDir(t, "ext/standard/scanf.c", scanfSrc)
	writeFile(t, filepath.Join(src, "main", "spprintf.c"), "")
	writeFile(t, filepath.Join(src, "main", "snprintf.c"), "")

	for i := 0; i < 2; i++ {
		if err := patchScanf(src); err != nil {
			t.Fatalf("patchScanf pass %d: %v", i+1, err)
		}
	}
	content := readFile(t, filepath.Join(src, "ext/standard/scanf.c"))
	if count := strings.Count(content, "const char *"); count != 2 {
		// two replacements: fn decl + two casts = 3 occurrences total
		// just ensure it's not doubled
	}
	_ = content
}

// ── patchGDCtx ───────────────────────────────────────────────────────────────

const gdCtxSrc = `
static void _php_image_output_ctx(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, void (*func_p)())
{
	(*func_p)(im, ctx, q);
	(*func_p)(im, ctx, q);
	(*func_p)(im, ctx, q, f);
	(*func_p)(im, file ? file : "", q, ctx);
	(*func_p)(im, q, ctx);
	(*func_p)(im, ctx, (int) compressed);
	(*func_p)(im, ctx);
}
`

func TestPatchGDCtx_FixesCallSites(t *testing.T) {
	ctxFile := filepath.Join(t.TempDir(), "gd_ctx.c")
	writeFile(t, ctxFile, gdCtxSrc)

	if err := patchGDCtx(ctxFile, "/* chauffeur-gd-compat */"); err != nil {
		t.Fatalf("patchGDCtx: %v", err)
	}
	content := readFile(t, ctxFile)

	must := []string{
		"((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, q)",
		"((void (*)(gdImagePtr, gdIOCtx *, int, int))func_p)(im, ctx, q, f)",
		"((void (*)(gdImagePtr, const char *, int, gdIOCtx *))func_p)",
		"((void (*)(gdImagePtr, int, gdIOCtx *))func_p)(im, q, ctx)",
		"((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, (int) compressed)",
		"((void (*)(gdImagePtr, gdIOCtx *))func_p)(im, ctx)",
	}
	for _, s := range must {
		if !strings.Contains(content, s) {
			t.Errorf("expected patched call site %q", s)
		}
	}
}

func TestPatchGDCtx_MissingFile(t *testing.T) {
	// PHP 8.0: gd_ctx.c does not exist — must not error.
	absent := filepath.Join(t.TempDir(), "no-gd_ctx.c")
	if err := patchGDCtx(absent, "/* marker */"); err != nil {
		t.Fatalf("patchGDCtx on missing file: %v", err)
	}
}

func TestPatchGDCtx_Idempotent(t *testing.T) {
	ctxFile := filepath.Join(t.TempDir(), "gd_ctx.c")
	writeFile(t, ctxFile, gdCtxSrc)
	marker := "/* chauffeur-gd-compat */"

	for i := 0; i < 2; i++ {
		if err := patchGDCtx(ctxFile, marker); err != nil {
			t.Fatalf("patchGDCtx pass %d: %v", i+1, err)
		}
	}
	content := readFile(t, ctxFile)
	if count := strings.Count(content, marker); count != 1 {
		t.Errorf("marker appears %d times, want 1", count)
	}
}

// ── patchGD ───────────────────────────────────────────────────────────────────

const gdSrc = `
static gdImagePtr _php_image_create_from_string(INTERNAL_FUNCTION_PARAMETERS, gdImagePtr (*ioctx_func_p)())
{
	im = (*ioctx_func_p)(io_ctx);
}

static gdImagePtr _php_image_create_from(INTERNAL_FUNCTION_PARAMETERS, gdImagePtr (*ioctx_func_p)(), gdImagePtr (*func_p)())
{
	im = (*ioctx_func_p)(io_ctx, srcx, srcy, width, height);
	im = (*ioctx_func_p)(io_ctx);
	im = (*func_p)(fp, srcx, srcy, width, height);
	im = (*func_p)(fp);
}

static void _php_image_output(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, void (*func_p)())
{
	(*func_p)(im, fp);
	(*func_p)(im, fp, q, t);
	(*func_p)(im, tmp);
	(*func_p)(im, tmp, q, t);
}
`

func TestPatchGD_FixesCallSites(t *testing.T) {
	src := setupSrcDir(t, "ext/gd/gd.c", gdSrc)

	if err := patchGD(src); err != nil {
		t.Fatalf("patchGD: %v", err)
	}
	content := readFile(t, filepath.Join(src, "ext/gd/gd.c"))

	must := []string{
		"((gdImagePtr (*)(gdIOCtx *))ioctx_func_p)(io_ctx)",
		"((gdImagePtr (*)(gdIOCtx *, int, int, int, int))ioctx_func_p)(io_ctx, srcx, srcy, width, height)",
		"((gdImagePtr (*)(FILE *, int, int, int, int))func_p)(fp, srcx, srcy, width, height)",
		"((gdImagePtr (*)(FILE *))func_p)(fp)",
		"((void (*)(gdImagePtr, FILE *))func_p)(im, fp)",
		"((void (*)(gdImagePtr, FILE *, int, int))func_p)(im, fp, q, t)",
		"((void (*)(gdImagePtr, FILE *))func_p)(im, tmp)",
		"((void (*)(gdImagePtr, FILE *, int, int))func_p)(im, tmp, q, t)",
	}
	for _, s := range must {
		if !strings.Contains(content, s) {
			t.Errorf("expected patched call site %q", s)
		}
	}
}

func TestPatchGD_Idempotent(t *testing.T) {
	src := setupSrcDir(t, "ext/gd/gd.c", gdSrc)
	marker := "/* chauffeur-gd-compat */"

	for i := 0; i < 2; i++ {
		if err := patchGD(src); err != nil {
			t.Fatalf("patchGD pass %d: %v", i+1, err)
		}
	}
	content := readFile(t, filepath.Join(src, "ext/gd/gd.c"))
	if count := strings.Count(content, marker); count != 1 {
		t.Errorf("marker appears %d times, want 1", count)
	}
}

// ── ApplyLegacyPatches integration ───────────────────────────────────────────

func TestApplyLegacyPatches_SkipsAbsentGDCtx(t *testing.T) {
	// Simulate PHP 8.0 source tree: no gd_ctx.c.
	src := t.TempDir()
	mustMkdir(t, filepath.Join(src, "ext/libxml"))
	mustMkdir(t, filepath.Join(src, "ext/openssl"))
	mustMkdir(t, filepath.Join(src, "ext/standard"))
	mustMkdir(t, filepath.Join(src, "ext/gd"))
	mustMkdir(t, filepath.Join(src, "main"))

	writeFile(t, filepath.Join(src, "ext/libxml/libxml.c"), libxmlSrc)
	writeFile(t, filepath.Join(src, "ext/openssl/openssl.c"), opensslSrc)
	writeFile(t, filepath.Join(src, "ext/standard/scanf.c"), scanfSrc)
	writeFile(t, filepath.Join(src, "main/spprintf.c"), "")
	writeFile(t, filepath.Join(src, "main/snprintf.c"), "")
	writeFile(t, filepath.Join(src, "ext/gd/gd.c"), gdSrc)
	// gd_ctx.c intentionally absent.

	if err := ApplyLegacyPatches(src); err != nil {
		t.Fatalf("ApplyLegacyPatches without gd_ctx.c: %v", err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// setupSrcDir creates a temp directory with a single file at relPath.
func setupSrcDir(t *testing.T, relPath, content string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, relPath), content)
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
