package installers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/siaji/chauffeur/cli/lib"
)

// applyLegacyPHPSourcePatches injects compatibility shims for legacy PHP releases
// that no longer build cleanly against modern system dependencies.
func applyLegacyPHPSourcePatches(version, sourceDir string, logger *lib.Logger) error {
	needsCompat := version == "7.4" || version == "8.0"
	if !needsCompat {
		return nil
	}

	if logger == nil {
		logger = lib.NewCommandLogger("install")
	}
	logPHPInfo(logger, "Applying compatibility patches for legacy PHP %s in directory %s", version, sourceDir)

	if err := patchLegacyLibxml(logger, sourceDir); err != nil {
		return err
	}
	if err := patchLegacyOpenSSL(logger, sourceDir); err != nil {
		return err
	}
	if err := patchLegacyScanf(logger, sourceDir); err != nil {
		return err
	}
	if err := patchLegacyGD(logger, sourceDir); err != nil {
		return err
	}

	logPHPSuccess(logger, "Legacy compatibility patches applied")
	return nil
}

func patchLegacyLibxml(logger *lib.Logger, sourceDir string) error {
	target := filepath.Join(sourceDir, "ext", "libxml", "libxml.c")
	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read libxml source: %w", err)
	}

	modified := false

	if !bytes.Contains(data, []byte("define ATTRIBUTE_UNUSED")) {
		snippet := []byte(`#ifndef ATTRIBUTE_UNUSED
# if defined(__GNUC__) || defined(__clang__)
#  define ATTRIBUTE_UNUSED __attribute__((unused))
# else
#  define ATTRIBUTE_UNUSED
# endif
#endif

`)

		logger.Info("Injecting ATTRIBUTE_UNUSED shim for libxml >= 2.12")
		logger.Info("Adjusting libxml handler casting for legacy build")
		data = append(snippet, data...)
		modified = true
	}

	replacements := map[string]string{
		"xmlSetStructuredErrorFunc(NULL, php_libxml_structured_error_handler);": "xmlSetStructuredErrorFunc(NULL, (xmlStructuredErrorFunc)php_libxml_structured_error_handler);",
		"current_handler == php_libxml_structured_error_handler":                "(xmlStructuredErrorFunc)current_handler == (xmlStructuredErrorFunc)php_libxml_structured_error_handler",
		"current_handler != php_libxml_structured_error_handler":                "(xmlStructuredErrorFunc)current_handler != (xmlStructuredErrorFunc)php_libxml_structured_error_handler",
	}

	for old, newVal := range replacements {
		if bytes.Contains(data, []byte(old)) && !bytes.Contains(data, []byte(newVal)) {
			logPHPInfo(logger, "Adjusting libxml handler casting for legacy build")
			updated := bytes.ReplaceAll(data, []byte(old), []byte(newVal))
			if !bytes.Equal(updated, data) {
				data = updated
				modified = true
			}
		}
	}

	if !modified {
		return nil
	}

	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write libxml shim: %w", err)
	}

	return nil
}

func patchLegacyOpenSSL(logger *lib.Logger, sourceDir string) error {
	target := filepath.Join(sourceDir, "ext", "openssl", "openssl.c")
	logPHPInfo(logger, "Patching OpenSSL file: %s", target)
	logPHPInfo(logger, "Adjusting EVP_PKEY RSA accessor for OpenSSL 3.x const correctness")
	logPHPInfo(logger, "Injecting OpenSSL 3.x compatibility shims")
	logPHPInfo(logger, "Patching OpenSSL shims in: %s", target)

	// Check if file exists before trying to read
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("openssl source file not found: %s", target)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read openssl source: %w", err)
	}

	modified := false

	// Always update call sites before inserting helper to ensure consistency.
	replaced := bytes.ReplaceAll(data, []byte("EVP_PKEY_get0_RSA("), []byte("EVP_PKEY_get0_RSA_NONCONST("))
	if !bytes.Equal(replaced, data) {
		logPHPInfo(logger, "Adjusting EVP_PKEY RSA accessor for OpenSSL 3.x const correctness")
		data = replaced
		modified = true
	}

	const compatMarker = "/* OpenSSL 3.x compatibility shims for legacy PHP */"
	if !bytes.Contains(data, []byte(compatMarker)) {
		compatSnippet := []byte(compatMarker + "\n" +
			"#include <openssl/opensslv.h>\n" +
			"#include <openssl/evp.h>\n" +
			"#include <openssl/rsa.h>\n" +
			"\n" +
			"#ifndef RSA_SSLV23_PADDING\n" +
			"#define RSA_SSLV23_PADDING 2\n" +
			"#endif\n" +
			"\n" +
			"#ifndef EVP_PKEY_get0_RSA_NONCONST\n" +
			"#define EVP_PKEY_get0_RSA_NONCONST(pkey) ((RSA*)EVP_PKEY_get0_RSA((pkey)))\n" +
			"#endif\n" +
			"/* ==== end OpenSSL 3.x shims ==== */\n\n",
		)
		logPHPInfo(logger, "Injecting OpenSSL 3.x compatibility shims")
		insertPos := findOpenSSLShimInsertPos(data)
		if insertPos < 0 {
			logger.Warn("could not locate config.h block; prepending shim", "")
			insertPos = 0
		}
		data = insertSnippet(data, compatSnippet, insertPos)
		if !bytes.Contains(data, []byte("EVP_PKEY_get0_RSA_NONCONST")) ||
			!bytes.Contains(data, []byte("RSA_SSLV23_PADDING")) {
			return fmt.Errorf("legacy OpenSSL shim missing after patch in %s", target)
		}
		logPHPInfo(logger, "Patching OpenSSL shims in: %s", target)
		modified = true
	}

	head := 80
	lines := 0
	var out bytes.Buffer
	for i := 0; i < len(data) && lines < head; i++ {
		out.WriteByte(data[i])
		if data[i] == '\n' {
			lines++
		}
	}
	logPHPInfo(logger, "openssl.c head:\n%s", out.String())

	if !bytes.Contains(data, []byte(compatMarker)) || !bytes.Contains(data, []byte("EVP_PKEY_get0_RSA_NONCONST")) {
		return fmt.Errorf("failed to inject OpenSSL compatibility shim into %s", target)
	}

	if !modified {
		return nil
	}

	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write openssl shim: %w", err)
	}

	return nil
}

func insertAfterOpenSSLInclude(data []byte, snippet []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 {
		return append(snippet, data...)
	}

	for i, line := range lines {
		if bytes.Equal(bytes.TrimSpace(line), []byte("#include <openssl/opensslv.h>")) {
			var builder bytes.Buffer
			endsWithNewline := len(data) > 0 && data[len(data)-1] == '\n'
			for idx, l := range lines {
				if idx > 0 {
					builder.WriteByte('\n')
				}
				builder.Write(l)
				if idx == i {
					builder.WriteByte('\n')
					builder.Write(snippet)
				}
			}
			if endsWithNewline {
				builder.WriteByte('\n')
			}
			return builder.Bytes()
		}
	}

	return insertAfterIncludeBlock(data, snippet)
}

func insertAfterIncludeBlock(data []byte, snippet []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) == 0 {
		return append(snippet, data...)
	}

	insertIndex := -1
	seenInclude := false
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("#include")) {
			seenInclude = true
			insertIndex = i + 1
			continue
		}
		if seenInclude {
			break
		}
	}

	if insertIndex == -1 {
		return append(snippet, data...)
	}

	var builder bytes.Buffer
	endsWithNewline := len(data) > 0 && data[len(data)-1] == '\n'
	for i, line := range lines {
		if i > 0 {
			builder.WriteByte('\n')
		}
		builder.Write(line)
		if i == insertIndex-1 {
			builder.WriteByte('\n')
			builder.Write(snippet)
		}
	}
	if endsWithNewline {
		builder.WriteByte('\n')
	}

	return builder.Bytes()
}

func findOpenSSLShimInsertPos(data []byte) int {
	if idx := bytes.Index(data, []byte("#include \"config.h\"")); idx >= 0 {
		rest := data[idx:]
		if endifIdx := bytes.Index(rest, []byte("#endif")); endifIdx >= 0 {
			pos := idx + endifIdx + len("#endif")
			for pos < len(data) && (data[pos] == '\n' || data[pos] == '\r') {
				pos++
			}
			return pos
		}
	}
	return insertAfterIncludeBlockPos(data)
}

func insertAfterIncludeBlockPos(data []byte) int {
	lines := bytes.Split(data, []byte("\n"))
	pos := 0
	for _, line := range lines {
		trim := bytes.TrimSpace(line)
		if bytes.HasPrefix(trim, []byte("#include")) {
			pos += len(line) + 1
			continue
		}
		break
	}
	return pos
}

func insertSnippet(data []byte, snippet []byte, pos int) []byte {
	if pos < 0 || pos > len(data) {
		pos = len(data)
	}
	var builder bytes.Buffer
	builder.Write(data[:pos])
	if pos > 0 && data[pos-1] != '\n' {
		builder.WriteByte('\n')
	}
	builder.Write(snippet)
	if pos < len(data) && data[pos] != '\n' {
		builder.WriteByte('\n')
	}
	builder.Write(data[pos:])
	return builder.Bytes()
}

func patchLegacyScanf(logger *lib.Logger, sourceDir string) error {
	target := filepath.Join(sourceDir, "ext", "standard", "scanf.c")
	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read scanf source: %w", err)
	}

	modified := false

	oldDecl := []byte("zend_long (*fn)() = NULL;")
	newDecl := []byte("zend_long (*fn)(const char *, char **, int) = NULL;")
	if bytes.Contains(data, oldDecl) && !bytes.Contains(data, newDecl) {
		logPHPInfo(logger, "Updating sscanf function pointer prototype for modern compilers")
		data = bytes.ReplaceAll(data, oldDecl, newDecl)
		modified = true
	}

	oldStrtol := []byte("fn = (zend_long (*)())ZEND_STRTOL_PTR;")
	newStrtol := []byte("fn = (zend_long (*)(const char *, char **, int))ZEND_STRTOL_PTR;")
	if bytes.Contains(data, oldStrtol) && !bytes.Contains(data, newStrtol) {
		data = bytes.ReplaceAll(data, oldStrtol, newStrtol)
		modified = true
	}

	oldStrtoul := []byte("fn = (zend_long (*)())ZEND_STRTOUL_PTR;")
	newStrtoul := []byte("fn = (zend_long (*)(const char *, char **, int))ZEND_STRTOUL_PTR;")
	if bytes.Contains(data, oldStrtoul) && !bytes.Contains(data, newStrtoul) {
		data = bytes.ReplaceAll(data, oldStrtoul, newStrtoul)
		modified = true
	}

	if !modified {
		return nil
	}

	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("write scanf shim: %w", err)
	}

	return nil
}

func patchLegacyGD(logger *lib.Logger, sourceDir string) error {
	logPHPInfo(logger, "Adjusting GD function pointers for newer libgd prototypes")

	gdFile := filepath.Join(sourceDir, "ext", "gd", "gd.c")
	ctxFile := filepath.Join(sourceDir, "ext", "gd", "gd_ctx.c")

	replaceOnce := func(data []byte, old, new string) ([]byte, error) {
		if !bytes.Contains(data, []byte(old)) {
			return data, fmt.Errorf("pattern %q not found", old)
		}
		return bytes.Replace(data, []byte(old), []byte(new), 1), nil
	}

	replaceAll := func(data []byte, old, new string) ([]byte, error) {
		if !bytes.Contains(data, []byte(old)) {
			return data, fmt.Errorf("pattern %q not found", old)
		}
		return bytes.ReplaceAll(data, []byte(old), []byte(new)), nil
	}

	gdData, err := os.ReadFile(gdFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", gdFile, err)
	}

	// Add a compatibility header at the top to handle all GD function signature mismatches
	compatHeader := []byte(`
/* GD compatibility fixes for modern compilers */
#ifdef __cplusplus
extern "C" {
#endif

/* Fix function signature mismatches - match what PHP 7.4 expects */
typedef void (*gd_output_func_t)(gdImagePtr, FILE *);
typedef void (*gd_output_func_ctx_t)(gdImagePtr, gdIOCtx *);
typedef void (*gd_output_func_int_t)(gdImagePtr, FILE *, int);
typedef void (*gd_output_func_int2_t)(gdImagePtr, FILE *, int, int);
typedef gdImagePtr (*gd_input_func_t)(FILE *);
typedef gdImagePtr (*gd_input_func_4arg_t)(FILE *, int, int, int, int);
typedef void (*gd_output_ctx_quality_t)(gdImagePtr, gdIOCtx *, int);

/* Additional function pointer types for various GD functions */
typedef gdImagePtr (*gd_input_func_ctx_t)(gdIOCtx *);
typedef void (*gd_output_func_wbmp_t)(gdImagePtr, int, FILE *);
typedef void (*gd_output_func_wbmp_ctx_t)(gdImagePtr, int, gdIOCtx *);
typedef void (*gd_output_func_png_ex_t)(gdImagePtr, gdIOCtx *, int, int);

#ifdef __cplusplus
}
#endif

`)

	// Prepend the compatibility header after the includes
	if !bytes.Contains(gdData, []byte("/* GD compatibility fixes for modern compilers */")) {
		// Find a good place to insert the header - after the includes
		insertPos := findGDHeaderInsertPos(gdData)
		gdData = insertSnippet(gdData, compatHeader, insertPos)
		logger.Info("Added GD compatibility header")
	}

	// Now fix function signatures to use the typedefs
	gdRepls := []struct {
		old string
		new string
		all bool
	}{
		// Fix function signature declarations - the patterns that exist in the code
		{
			old: "static void _php_image_create_from(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, gdImagePtr (*func_p)(), gdImagePtr (*ioctx_func_p)())",
			new: "static void _php_image_create_from(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, gd_input_func_t func_p, gd_input_func_t ioctx_func_p)",
			all: false,
		},
		{
			old: "static void _php_image_create_from_string(zval *data, char *tn, gdImagePtr (*ioctx_func_p)())",
			new: "static void _php_image_create_from_string(zval *data, char *tn, gd_input_func_t ioctx_func_p)",
			all: false,
		},
		{
			old: "static void _php_image_output(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, void (*func_p)())",
			new: "static void _php_image_output(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, gd_output_func_t func_p)",
			all: false,
		},
		{
			old: "static void _php_image_output_ctx(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, void (*func_p)())",
			new: "static void _php_image_output_ctx(INTERNAL_FUNCTION_PARAMETERS, int image_type, char *tn, gd_output_func_ctx_t func_p)",
			all: false,
		},
		// Fix function calls to use proper casting
		{
			old: "im = (*func_p)(io_ctx);",
			new: "im = ((gd_input_func_t)func_p)(io_ctx);",
			all: true,
		},
		{
			old: "im = (*func_p)(fp);",
			new: "im = ((gd_input_func_t)func_p)(fp);",
			all: true,
		},
		{
			old: "im = (*func_p)(fp, srcx, srcy, width, height);",
			new: "im = ((gd_input_func_4arg_t)func_p)(fp, srcx, srcy, width, height);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp);",
			new: "((gd_output_func_t)func_p)(im, fp);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp, q);",
			new: "((gd_output_func_int_t)func_p)(im, fp, q);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp, q, t);",
			new: "((gd_output_func_int2_t)func_p)(im, fp, q, t);",
			all: true,
		},
		{
			old: "(*func_p)(im, ctx);",
			new: "((gd_output_func_ctx_t)func_p)(im, ctx);",
			all: true,
		},
		{
			old: "(*func_p)(im, ctx, (int) quality);",
			new: "((gd_output_func_ctx_t)func_p)(im, ctx, (int) quality);",
			all: true,
		},
		{
			old: "(*func_p)(im, tmp);",
			new: "((gd_output_func_t)func_p)(im, tmp);",
			all: true,
		},
	}

	for _, repl := range gdRepls {
		if repl.all {
			gdData, err = replaceAll(gdData, repl.old, repl.new)
		} else {
			gdData, err = replaceOnce(gdData, repl.old, repl.new)
		}
		if err != nil {
			// Log but don't fail if pattern not found (different PHP versions may have slightly different code)
			logger.Warn("GD patch pattern not found", err.Error())
		}
	}

	if err := os.WriteFile(gdFile, gdData, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", gdFile, err)
	}

	// Check if gd_ctx.c exists before trying to patch it
	if _, err := os.Stat(ctxFile); os.IsNotExist(err) {
		logPHPInfo(logger, "Skipping gd_ctx.c adjustments (file not found in %s)", sourceDir)
		return nil
	}

	ctxData, err := os.ReadFile(ctxFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", ctxFile, err)
	}

	ctxRepls := []struct {
		old string
		new string
		all bool
	}{
		{
			old: "(*func_p)(im, ctx, q);",
			new: "((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, q);",
			all: true,
		},
		{
			old: "(*func_p)(im, ctx, q, f);",
			new: "((void (*)(gdImagePtr, gdIOCtx *, int, int))func_p)(im, ctx, q, f);",
			all: true,
		},
		{
			old: "(*func_p)(im, file ? file : \"\", q, ctx);",
			new: "((void (*)(gdImagePtr, char *, int, gdIOCtx *))func_p)(im, file ? file : \"\", q, ctx);",
		},
		{
			old: "(*func_p)(im, q, ctx);",
			new: "((void (*)(gdImagePtr, int, gdIOCtx *))func_p)(im, q, ctx);",
			all: true,
		},
		{
			old: "(*func_p)(im, ctx, (int) compressed);",
			new: "((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, (int) compressed);",
		},
		{
			old: "(*func_p)(im, ctx);",
			new: "((void (*)(gdImagePtr, gdIOCtx *))func_p)(im, ctx);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp);",
			new: "((void (*)(gdImagePtr, FILE *))func_p)(im, fp);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp, q);",
			new: "((void (*)(gdImagePtr, FILE *, int))func_p)(im, fp, q);",
			all: true,
		},
		{
			old: "(*func_p)(im, fp, q, t);",
			new: "((void (*)(gdImagePtr, FILE *, int, int))func_p)(im, fp, q, t);",
			all: true,
		},
		{
			old: "(*func_p)(im, ctx, (int) quality);",
			new: "((void (*)(gdImagePtr, gdIOCtx *, int))func_p)(im, ctx, (int) quality);",
			all: true,
		},
	}

	for _, repl := range ctxRepls {
		if repl.all {
			ctxData, err = replaceAll(ctxData, repl.old, repl.new)
		} else {
			ctxData, err = replaceOnce(ctxData, repl.old, repl.new)
		}
		if err != nil {
			// Log but don't fail if pattern not found (different PHP versions may have slightly different code)
			logger.Warn("GD ctx patch pattern not found", err.Error())
		}
	}

	if err := os.WriteFile(ctxFile, ctxData, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ctxFile, err)
	}

	return nil
}

func findGDHeaderInsertPos(data []byte) int {
	// Find the end of the initial includes and defines block
	lines := bytes.Split(data, []byte("\n"))
	includeEnd := -1
	inIncludes := true

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)

		// Stop including when we hit function definitions or code
		if bytes.HasPrefix(trimmed, []byte("static void")) ||
		   bytes.HasPrefix(trimmed, []byte("PHP_FUNCTION")) ||
		   bytes.HasPrefix(trimmed, []byte("PHP_NAMED_FUNCTION")) {
			break
		}

		// Track if we're still in the includes/defines section
		if bytes.HasPrefix(trimmed, []byte("#include")) ||
		   bytes.HasPrefix(trimmed, []byte("#define")) ||
		   len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte("/*")) {
			includeEnd = i
			continue
		}

		// If we hit a non-include line, we're probably done with the header
		if inIncludes && !bytes.HasPrefix(trimmed, []byte("#")) &&
		   !bytes.HasPrefix(trimmed, []byte("/*")) &&
		   !bytes.HasPrefix(trimmed, []byte("typedef")) &&
		   len(trimmed) > 0 {
			inIncludes = false
		}
	}

	if includeEnd >= 0 {
		// Find the byte position for the end of that line
		pos := 0
		for i := 0; i <= includeEnd && i < len(lines); i++ {
			pos += len(lines[i]) + 1 // +1 for newline
		}
		return pos
	}

	// Fallback: insert after the first few include blocks
	return insertAfterIncludeBlockPos(data)
}
