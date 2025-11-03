package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/siaji/chauffeur/cli/installers"
)

func TestHandleChecksumVerifiedSHA512(t *testing.T) {
	path := createTempFile(t, "hello world")
	sha512Hex, err := installers.TestFileSHA512(path)
	if err != nil {
		t.Fatalf("fileSHA512: %v", err)
	}
	t.Setenv("CHAUF_REQUIRE_CHECKSUM", "0")
	output := captureOutput(func() error {
		return installers.TestHandleChecksum(path, sha512Hex, "sha512")
	})
	if !strings.Contains(output, "Upstream checksum verified (sha512)") {
		t.Fatalf("expected success output, got %q", output)
	}
}

func TestHandleChecksumNoChecksum(t *testing.T) {
	path := createTempFile(t, "checksumless content")
	t.Setenv("CHAUF_REQUIRE_CHECKSUM", "0")
	output := captureOutput(func() error {
		return installers.TestHandleChecksum(path, "", "")
	})
	if !strings.Contains(output, "No upstream checksum asset found") {
		t.Fatalf("expected warning output, got %q", output)
	}
	if !strings.Contains(output, "Local SHA256") || !strings.Contains(output, "Local SHA512") {
		t.Fatalf("expected local digest output, got %q", output)
	}
}

func TestHandleChecksumMismatch(t *testing.T) {
	path := createTempFile(t, "mismatch")
	badSum := strings.Repeat("a", 64)
	err := installers.TestHandleChecksum(path, badSum, "sha256")
	if err == nil {
		t.Fatal("expected error for mismatched checksum")
	}
	actual, _ := installers.TestFileSHA256(path)
	if !strings.Contains(err.Error(), actual) {
		t.Fatalf("expected actual checksum in error, got %v", err)
	}
}

func TestEvaluateFingerprintOptionalMismatch(t *testing.T) {
	key := installers.TestServiceKeys()[0]
	output := captureOutput(func() error {
		_, err := installers.TestEvaluateFingerprint("DEADBEEF", key.Fingerprint, key.Name, true)
		return err
	})
	if !strings.Contains(output, "Warning") {
		t.Fatalf("expected warning output, got %q", output)
	}
}

func TestEvaluateFingerprintMaintainerValid(t *testing.T) {
	key := installers.TestServiceKeys()[1]
	ok, err := installers.TestEvaluateFingerprint(strings.ToLower(key.Fingerprint), key.Fingerprint, key.Name, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected maintainer key to be accepted")
	}
}

func TestParseValidSignatures(t *testing.T) {
	key := installers.TestServiceKeys()[1]
	status := fmt.Sprintf("[GNUPG:] VALIDSIG %s 20250101T000000 0 4 0 1 8 01 %s\n", strings.ToLower(key.Fingerprint), strings.ToLower(key.Fingerprint))
	status += "[GNUPG:] GOODSIG 1234567890 roman arutyunyan <r.arutyunyan@f5.com>\n"
	if !installers.TestParseValidSignatures(status) {
		t.Fatalf("expected valid signature to be accepted")
	}

	if installers.TestParseValidSignatures("[GNUPG:] VALIDSIG 1234567890ABCDEF1234567890ABCDEF123456 0 0 0 0\n") {
		t.Fatalf("expected unknown fingerprint to fail")
	}
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "chauf-test-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}
