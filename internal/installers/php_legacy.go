package installers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// legacyVersions are PHP versions that require vendored OpenSSL and source patches.
var legacyVersions = map[string]bool{
	"7.4": true,
	"8.0": true,
}

const (
	openSSLVersion  = "1.1.1w"
	openSSLURL      = "https://www.openssl.org/source/openssl-1.1.1w.tar.gz"
	openSSLFallback = "https://www.openssl.org/source/old/1.1.1/openssl-1.1.1w.tar.gz"
	openSSLSHA256   = "cf3098950cb4d853ad95c0841f1f9c6d3dc102dccfcacd521d93925208b76ac8"
)

// IsLegacy reports whether majorMinor (e.g. "7.4", "8.0") requires legacy treatment.
func IsLegacy(majorMinor string) bool {
	return legacyVersions[majorMinor]
}

// BuildVendoredOpenSSL downloads and compiles OpenSSL 1.1.1w into vendorDir.
// Uses cacheDir to store the downloaded tarball.
func BuildVendoredOpenSSL(cacheDir, vendorDir string, opts BuildOpts) error {
	tarPath := filepath.Join(cacheDir, "openssl-"+openSSLVersion+".tar.gz")

	_, err := DownloadWithProgress(openSSLURL, tarPath, opts.NoCache)
	if err != nil {
		_, err = DownloadWithProgress(openSSLFallback, tarPath, opts.NoCache)
		if err != nil {
			return fmt.Errorf("download OpenSSL: %w", err)
		}
	}

	if err := VerifySHA256(tarPath, openSSLSHA256); err != nil {
		return fmt.Errorf("OpenSSL checksum: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "chauf-openssl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	srcDir, err := ExtractTarGz(tarPath, tmpDir, 0)
	if err != nil {
		return fmt.Errorf("extract OpenSSL: %w", err)
	}

	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		return err
	}

	configureArgs := []string{
		"--prefix=" + vendorDir,
		"--openssldir=" + vendorDir,
		"shared",
		"zlib",
		"-fPIC",
	}

	if err := RunCmd(srcDir, opts.Verbose, "./config", configureArgs...); err != nil {
		return fmt.Errorf("OpenSSL config: %w", err)
	}
	if err := RunCmd(srcDir, opts.Verbose, "make", "-j"+NumCPUs()); err != nil {
		return fmt.Errorf("OpenSSL make: %w", err)
	}
	// install_sw skips man pages and speeds things up
	if err := RunCmd(srcDir, opts.Verbose, "make", "install_sw"); err != nil {
		return fmt.Errorf("OpenSSL install: %w", err)
	}
	return nil
}

// VendoredOpenSSLEnv returns env vars that point the PHP build at a vendored OpenSSL.
// -Wno-incompatible-pointer-types is required for GCC 14+ where PHP 7.4/8.0's GD
// extension uses old-style void (*func_p)() function pointers that are now errors.
func VendoredOpenSSLEnv(vendorDir string) []string {
	return []string{
		"CFLAGS=-Wno-deprecated-declarations -Wno-discarded-qualifiers -Wno-incompatible-pointer-types",
		"CPPFLAGS=-DOPENSSL_API_COMPAT=0x10100000L",
		"PKG_CONFIG_PATH=" + filepath.Join(vendorDir, "lib", "pkgconfig"),
		"LD_LIBRARY_PATH=" + filepath.Join(vendorDir, "lib"),
	}
}

// ApplyLegacyPatches applies the source patches needed for PHP 7.4 and 8.0.
func ApplyLegacyPatches(srcDir string) error {
	if err := patchLibXML(srcDir); err != nil {
		return fmt.Errorf("libxml patch: %w", err)
	}
	if err := patchOpenSSLExt(srcDir); err != nil {
		return fmt.Errorf("openssl ext patch: %w", err)
	}
	if err := patchScanf(srcDir); err != nil {
		return fmt.Errorf("scanf patch: %w", err)
	}
	if err := patchGD(srcDir); err != nil {
		return fmt.Errorf("gd patch: %w", err)
	}
	return nil
}

// patchLibXML fixes three compatibility issues in ext/libxml/libxml.c:
//  1. ATTRIBUTE_UNUSED macro — the identifier is used in the original source but
//     not defined by modern libxml2 headers; we inject the #define.
//  2. xmlOutputBufferCreateFilenameDefault — removed in libxml2 2.12; guarded with
//     a version check so the build still works on older distros.
//  3. xmlStructuredSetErrors handler cast — required for strict pointer type checking.
func patchLibXML(srcDir string) error {
	path := filepath.Join(srcDir, "ext", "libxml", "libxml.c")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// Fix 1: inject ATTRIBUTE_UNUSED macro if not already defined.
	// Important: check for "#define ATTRIBUTE_UNUSED", not just the identifier,
	// because the original PHP 7.4 source uses the identifier on line ~479 without
	// defining it (expecting system headers to define it — but modern libxml2 doesn't).
	const macro = "#ifndef ATTRIBUTE_UNUSED\n#define ATTRIBUTE_UNUSED __attribute__((unused))\n#endif\n\n"
	if !strings.Contains(content, "#define ATTRIBUTE_UNUSED") {
		content = macro + content
	}

	// Fix 2: guard xmlOutputBufferCreateFilenameDefault calls removed in libxml2 2.12.
	// The function and its callback type were dropped; wrap both call sites.
	const oldBufCall = "xmlOutputBufferCreateFilenameDefault(php_libxml_output_buffer_create_filename);"
	const newBufCall = "#if LIBXML_VERSION < 21200\n\t\t\t\txmlOutputBufferCreateFilenameDefault(php_libxml_output_buffer_create_filename);\n\t\t\t\t#endif"
	content = strings.ReplaceAll(content, oldBufCall, newBufCall)

	// Fix 3: cast handler where it is passed as a function pointer argument.
	// Apply line-by-line to avoid corrupting the function definition line.
	const handler = "php_libxml_structured_error_handler"
	const cast = "(xmlStructuredErrorFunc)php_libxml_structured_error_handler"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, handler) && !strings.Contains(line, cast) {
			// Skip declaration and definition lines
			if strings.Contains(line, "void "+handler) || strings.Contains(line, "static "+handler) {
				continue
			}
			lines[i] = strings.ReplaceAll(line, handler, cast)
		}
	}
	content = strings.Join(lines, "\n")

	return os.WriteFile(path, []byte(content), 0644)
}

// patchOpenSSLExt injects OpenSSL 3.x compatibility shims and fixes
// EVP_PKEY_get0_RSA calls in ext/openssl/openssl.c.
func patchOpenSSLExt(srcDir string) error {
	path := filepath.Join(srcDir, "ext", "openssl", "openssl.c")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// Step 1: replace existing EVP_PKEY_get0_RSA( calls before injecting the shim
	// (so the shim macro definition uses the original name).
	if !strings.Contains(content, "EVP_PKEY_get0_RSA_NONCONST") {
		content = strings.ReplaceAll(content, "EVP_PKEY_get0_RSA(", "EVP_PKEY_get0_RSA_NONCONST(")
	}

	// Step 2: inject compat block right after the config.h include.
	const configInclude = `#ifdef HAVE_CONFIG_H
#include "config.h"
#endif`
	const shims = `/* OpenSSL 3.x compatibility shims for legacy PHP */
#ifndef RSA_SSLV23_PADDING
#define RSA_SSLV23_PADDING 2
#endif
#ifndef EVP_PKEY_get0_RSA_NONCONST
#define EVP_PKEY_get0_RSA_NONCONST(pkey) ((RSA*)EVP_PKEY_get0_RSA((pkey)))
#endif`

	if !strings.Contains(content, "OpenSSL 3.x compatibility") && strings.Contains(content, configInclude) {
		content = strings.Replace(content, configInclude, configInclude+"\n\n"+shims, 1)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// patchScanf fixes zend_long function pointer prototypes in two places:
//
//  1. main/spprintf.c and main/snprintf.c — generic (*fn)() → (*fn)(void *)
//  2. ext/standard/scanf.c — typed strtol/strtoul dispatch with concrete signature
//
// Both are the same GCC 14 empty-parens issue: `()` now means `(void)` in C99.
func patchScanf(srcDir string) error {
	// ── main/spprintf.c and main/snprintf.c ──────────────────────────────
	for _, rel := range []string{"main/spprintf.c", "main/snprintf.c"} {
		path := filepath.Join(srcDir, rel)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		content := strings.ReplaceAll(string(data),
			"zend_long (*fn)()",
			"zend_long (*fn)(void *)",
		)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	// ── ext/standard/scanf.c ─────────────────────────────────────────────
	// This file declares `zend_long (*fn)() = NULL` and then assigns either
	// ZEND_STRTOL_PTR or ZEND_STRTOUL_PTR before calling (*fn)(buf, NULL, base).
	// We fix the declaration to carry the concrete signature and update the casts.
	scanfPath := filepath.Join(srcDir, "ext", "standard", "scanf.c")
	data, err := os.ReadFile(scanfPath)
	if err != nil {
		return err
	}
	content := string(data)

	content = strings.ReplaceAll(content,
		"zend_long (*fn)() = NULL;",
		"zend_long (*fn)(const char *, char **, int) = NULL;")
	content = strings.ReplaceAll(content,
		"fn = (zend_long (*)())ZEND_STRTOL_PTR;",
		"fn = (zend_long (*)(const char *, char **, int))ZEND_STRTOL_PTR;")
	content = strings.ReplaceAll(content,
		"fn = (zend_long (*)())ZEND_STRTOUL_PTR;",
		"fn = (zend_long (*)(const char *, char **, int))ZEND_STRTOUL_PTR;")

	return os.WriteFile(scanfPath, []byte(content), 0644)
}

// patchGDCtx patches gd_ctx.c call sites. It is a no-op if the file does not
// exist (PHP 8.0 merged gd_ctx.c into gd.c).
func patchGDCtx(ctxFile, marker string) error {
	data, err := os.ReadFile(ctxFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	ctx := string(data)
	if strings.Contains(ctx, marker) {
		return nil
	}
	ctx = marker + "\n" + ctx
	// Order matters: patch more-specific patterns before shorter ones.
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, ctx, q, f);",
		"((void (*)(gdImagePtr, gdIOCtx *, int, int))func_p)(im, ctx, q, f);")
	ctx = strings.ReplaceAll(ctx,
		`(*func_p)(im, file ? file : "", q, ctx);`,
		`((void (*)(gdImagePtr, const char *, int, gdIOCtx *))func_p)(im, file ? file : "", q, ctx);`)
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, q, ctx);",
		"((void (*)(gdImagePtr, int, gdIOCtx *))func_p)(im, q, ctx);")
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, ctx, (int) compressed);",
		"((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, (int) compressed);")
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, ctx, (int) quality);",
		"((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, (int) quality);")
	// Generic 3-arg ctx call — must come after the more specific ones above.
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, ctx, q);",
		"((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, q);")
	ctx = strings.ReplaceAll(ctx,
		"(*func_p)(im, ctx);",
		"((void (*)(gdImagePtr, gdIOCtx *))func_p)(im, ctx);")
	return os.WriteFile(ctxFile, []byte(ctx), 0644)
}

// patchGD fixes GD extension function pointer call sites for GCC 14+.
//
// PHP 7.4's GD code declares internal callbacks as void (*func_p)(), which in
// C99/C11 means "takes no arguments". GCC 14 enforces this strictly, so every
// call site like (*func_p)(im, ctx, q) becomes a hard error ("too many arguments").
//
// The fix: cast each call site to the concrete function pointer type so the
// compiler knows the actual calling convention. We also keep
// -Wno-incompatible-pointer-types in CFLAGS to suppress the assignment-level
// warnings where typed function pointers are passed as void(*func_p)().
func patchGD(srcDir string) error {
	const marker = "/* chauffeur-gd-compat */"

	// ── gd_ctx.c — _php_image_output_ctx call sites ─────────────────────
	// PHP 8.0 merged gd_ctx.c into gd.c; skip if absent.
	if err := patchGDCtx(filepath.Join(srcDir, "ext", "gd", "gd_ctx.c"), marker); err != nil {
		return err
	}

	// ── gd.c — _php_image_create_from* and _php_image_output call sites ──
	gdFile := filepath.Join(srcDir, "ext", "gd", "gd.c")
	gdData, err := os.ReadFile(gdFile)
	if err != nil {
		return err
	}
	gd := string(gdData)

	if !strings.Contains(gd, marker) {
		gd = marker + "\n" + gd

		// ioctx_func_p — returns gdImagePtr, takes gdIOCtx * (optionally + crop args)
		gd = strings.ReplaceAll(gd,
			"im = (*ioctx_func_p)(io_ctx, srcx, srcy, width, height);",
			"im = ((gdImagePtr (*)(gdIOCtx *, int, int, int, int))ioctx_func_p)(io_ctx, srcx, srcy, width, height);")
		gd = strings.ReplaceAll(gd,
			"im = (*ioctx_func_p)(io_ctx);",
			"im = ((gdImagePtr (*)(gdIOCtx *))ioctx_func_p)(io_ctx);")

		// func_p (FILE *-based) — returns gdImagePtr or void
		gd = strings.ReplaceAll(gd,
			"im = (*func_p)(fp, srcx, srcy, width, height);",
			"im = ((gdImagePtr (*)(FILE *, int, int, int, int))func_p)(fp, srcx, srcy, width, height);")
		gd = strings.ReplaceAll(gd,
			"im = (*func_p)(fp);",
			"im = ((gdImagePtr (*)(FILE *))func_p)(fp);")

		// void output variants — 4-arg before 2-arg to avoid prefix collision
		gd = strings.ReplaceAll(gd,
			"(*func_p)(im, fp, q, t);",
			"((void (*)(gdImagePtr, FILE *, int, int))func_p)(im, fp, q, t);")
		gd = strings.ReplaceAll(gd,
			"(*func_p)(im, tmp, q, t);",
			"((void (*)(gdImagePtr, FILE *, int, int))func_p)(im, tmp, q, t);")
		gd = strings.ReplaceAll(gd,
			"(*func_p)(im, fp);",
			"((void (*)(gdImagePtr, FILE *))func_p)(im, fp);")
		gd = strings.ReplaceAll(gd,
			"(*func_p)(im, tmp);",
			"((void (*)(gdImagePtr, FILE *))func_p)(im, tmp);")

		if err := os.WriteFile(gdFile, []byte(gd), 0644); err != nil {
			return err
		}
	}

	return nil
}
