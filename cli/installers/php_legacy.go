package installers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// applyLegacyPHPSourcePatches injects compatibility shims for legacy PHP releases
// that no longer build cleanly against modern system dependencies.
func applyLegacyPHPSourcePatches(version, sourceDir string) error {
	needsCompat := version == "7.4" || version == "8.0"
	if !needsCompat {
		return nil
	}

	logPHPInfo("Applying compatibility patches for legacy PHP %s", version)

	if err := patchLegacyLibxml(sourceDir); err != nil {
		return err
	}
	if err := patchLegacyOpenSSL(sourceDir); err != nil {
		return err
	}

	logPHPSuccess("Legacy compatibility patches applied")
	return nil
}

func patchLegacyLibxml(sourceDir string) error {
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

		logPHPInfo("Injecting ATTRIBUTE_UNUSED shim for libxml >= 2.12")
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
			logPHPInfo("Adjusting libxml handler casting for legacy build")
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

func patchLegacyOpenSSL(sourceDir string) error {
	target := filepath.Join(sourceDir, "ext", "openssl", "openssl.c")
	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read openssl source: %w", err)
	}

	modified := false

	// Always update call sites before inserting helper to ensure consistency.
	replaced := bytes.ReplaceAll(data, []byte("EVP_PKEY_get0_RSA("), []byte("EVP_PKEY_get0_RSA_NONCONST("))
	if !bytes.Equal(replaced, data) {
		logPHPInfo("Adjusting EVP_PKEY RSA accessor for OpenSSL 3.x const correctness")
		data = replaced
		modified = true
	}

	if !bytes.Contains(data, []byte("#define RSA_SSLV23_PADDING")) {
		snippet := []byte("#ifndef RSA_SSLV23_PADDING\n" +
			"#define RSA_SSLV23_PADDING 2\n" +
			"#endif\n\n")
		logPHPInfo("Injecting RSA_SSLV23_PADDING shim for OpenSSL 3.x")
		data = insertAfterOpenSSLInclude(data, snippet)
		modified = true
	}

	if !bytes.Contains(data, []byte("#define EVP_PKEY_get0_RSA_NONCONST")) {
		snippet := []byte("#ifndef EVP_PKEY_get0_RSA_NONCONST\n" +
			"#include <stdint.h>\n" +
			"#include <openssl/evp.h>\n" +
			"#include <openssl/rsa.h>\n" +
			"#define EVP_PKEY_get0_RSA_NONCONST(k) \\\n" +
			"    ((RSA *)(uintptr_t)(const void *)EVP_PKEY_get0_RSA((k)))\n" +
			"#endif\n\n")

		logPHPInfo("Injecting EVP_PKEY_get0_RSA_NONCONST helper for legacy build")
		data = insertAfterOpenSSLInclude(data, snippet)
		modified = true
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
