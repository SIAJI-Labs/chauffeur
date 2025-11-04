package lib

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ChecksumFromList downloads a checksum list and extracts the value for assetName.
//
// @param client HTTP client used to fetch the list.
// @param url    Location of the checksum manifest.
// @param assetName File name whose checksum is required.
// @return Matching checksum string for the asset.
func ChecksumFromList(client *http.Client, url, assetName string) (string, error) {
	content, err := DownloadText(client, url)
	if err != nil {
		return "", err
	}
	return ChecksumFromContent(content, assetName)
}

// ChecksumFromContent parses checksum text and returns the hash for assetName.
//
// @param content  Raw checksum list contents.
// @param assetName File name to match against the list.
// @return Checksum string matching assetName.
func ChecksumFromContent(content, assetName string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		fieldCount := len(fields)

		var sum, name string
		if fieldCount == 1 {
			sum = fields[0]
			name = ""
		} else {
			sum = fields[0]
			name = strings.TrimPrefix(fields[len(fields)-1], "*")
		}

		if assetName == "" || name == assetName || fieldCount == 1 {
			return sum, nil
		}
	}

	if assetName == "" {
		return "", fmt.Errorf("checksum not found in provided content")
	}
	return "", fmt.Errorf("checksum for %s not found in provided content", assetName)
}

// HandleChecksum validates a downloaded file against expected checksums.
// It supports both SHA256 and SHA512 algorithms.
//
// @param filePath Path to the file to validate.
// @param expected Expected checksum value (can be a single checksum or multiple).
// @return Valid checksum string or an error if validation fails.
func HandleChecksum(filePath string, expected ...string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Try SHA256 first
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	sha256Sum := hex.EncodeToString(hasher.Sum(nil))

	// Check if any expected checksum matches SHA256
	for _, exp := range expected {
		if exp == sha256Sum {
			return sha256Sum, nil
		}
	}

	// If SHA256 didn't match, try SHA512
	file.Seek(0, 0)
	hasher = sha512.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	sha512Sum := hex.EncodeToString(hasher.Sum(nil))

	// Check if any expected checksum matches SHA512
	for _, exp := range expected {
		if exp == sha512Sum {
			return sha512Sum, nil
		}
	}

	return "", fmt.Errorf("checksum mismatch: got SHA256=%s, SHA512=%s, expected one of: %v", sha256Sum, sha512Sum, expected)
}

// FileHash returns the SHA256 hash of a file.
//
// @param filePath Path to the file to hash.
// @return Hexadecimal SHA256 hash.
func FileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
