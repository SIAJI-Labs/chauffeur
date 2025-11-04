package installers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/siaji/chauffeur/cli/internal/logging"
)

// applyLegacyPHPSourcePatches injects compatibility shims for legacy PHP releases
// that no longer build cleanly against modern system dependencies.
func applyLegacyPHPSourcePatches(version, sourceDir string, logger *logging.CommandLogger) error {
	needsCompat := version == "7.4" || version == "8.0"
	if !needsCompat {
		return nil
	}

	if logger == nil {
		logger = logging.NewCommandLogger("install")
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

	logPHPSuccess(logger, "Legacy compatibility patches applied")
	return nil
}

func patchLegacyLibxml(logger *logging.CommandLogger, sourceDir string) error {
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

func patchLegacyOpenSSL(logger *logging.CommandLogger, sourceDir string) error {
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

func patchLegacyScanf(logger *logging.CommandLogger, sourceDir string) error {
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
