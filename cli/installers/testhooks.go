package installers

import "github.com/siaji/chauffeur/cli/internal/logging"

// TestSigningKey mirrors minimal signing key metadata for tests.
type TestSigningKey struct {
	Name        string
	Fingerprint string
	Optional    bool
}

// TestHandleChecksum exposes handleChecksum for external tests.
func TestHandleChecksum(path, sum, algo string) error {
	return handleChecksum(path, sum, algo, logging.NewCommandLogger("test"))
}

// TestAlgorithmFromChecksum exposes algorithmFromChecksum for external tests.
func TestAlgorithmFromChecksum(sum, defaultAlgo string) string {
	return algorithmFromChecksum(sum, defaultAlgo)
}

// TestEvaluateFingerprint exposes evaluateFingerprint for external tests.
func TestEvaluateFingerprint(actual, expected, name string, optional bool) (bool, error) {
	return evaluateFingerprint(actual, expected, name, optional)
}

// TestParseValidSignatures exposes parseValidSignatures for external tests.
func TestParseValidSignatures(status string) bool {
	return parseValidSignatures(status)
}

// TestIsHexString exposes isHexString for external tests.
func TestIsHexString(s string) bool {
	return isHexString(s)
}

// TestFileSHA256 exposes fileSHA256 for external tests.
func TestFileSHA256(path string) (string, error) {
	return fileSHA256(path)
}

// TestFileSHA512 exposes fileSHA512 for external tests.
func TestFileSHA512(path string) (string, error) {
	return fileSHA512(path)
}

// TestServiceKeys provides a copy of known signing keys.
func TestServiceKeys() []TestSigningKey {
	out := make([]TestSigningKey, len(nginxSigningKeys))
	for i, k := range nginxSigningKeys {
		out[i] = TestSigningKey{Name: k.Name, Fingerprint: k.Fingerprint, Optional: k.Optional}
	}
	return out
}
