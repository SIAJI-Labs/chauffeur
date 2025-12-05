package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var githubAPIBase = "https://api.github.com"
var clientFactory = func() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// setGitHubAPIBase allows tests to override the GitHub API host.
func setGitHubAPIBase(base string) {
	githubAPIBase = base
}

// GitHubRelease represents a GitHub API response for a release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// GetLatestVersion fetches the latest release version from GitHub
func GetLatestVersion() (string, error) {
	client := clientFactory()

	resp, err := client.Get(fmt.Sprintf("%s/repos/SIAJI-Labs/chauffeur/releases/latest", githubAPIBase))
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	// Skip draft or pre-release versions
	if release.Draft || release.Prerelease {
		return GetLatestReleaseVersion()
	}

	return release.TagName, nil
}

// GetLatestReleaseVersion finds the latest non-prerelease, non-draft release
func GetLatestReleaseVersion() (string, error) {
	client := clientFactory()

	resp, err := client.Get(fmt.Sprintf("%s/repos/SIAJI-Labs/chauffeur/releases", githubAPIBase))
	if err != nil {
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	for _, release := range releases {
		if !release.Draft && !release.Prerelease && release.TagName != "" {
			return release.TagName, nil
		}
	}

	return "", fmt.Errorf("no stable releases found")
}

// CompareVersions compares two version strings (removing 'v' prefix if present)
// Returns: -1 if current < latest, 0 if equal, 1 if current > latest
func CompareVersions(current, latest string) int {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Simple version comparison - for semantic versioning like 1.2.3
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		var currentNum, latestNum int

		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &currentNum)
		}
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &latestNum)
		}

		if currentNum < latestNum {
			return -1
		} else if currentNum > latestNum {
			return 1
		}
	}

	return 0
}

// IsUpdateAvailable checks if there's a newer version available
func IsUpdateAvailable(currentVersion string) (bool, string, error) {
	latest, err := GetLatestVersion()
	if err != nil {
		return false, "", err
	}

	comparison := CompareVersions(currentVersion, latest)
	isAvailable := comparison < 0

	return isAvailable, latest, nil
}
