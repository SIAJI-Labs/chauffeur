package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GitHubRelease models the subset of release data we care about.
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a single downloadable artifact.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// LatestGitHubRelease fetches the latest published release metadata.
func LatestGitHubRelease(client *http.Client, owner, repo string) (GitHubRelease, error) {
	if client == nil {
		client = &http.Client{}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "chauffeur-cli")

	resp, err := client.Do(req)
	if err != nil {
		return GitHubRelease{}, fmt.Errorf("fetch latest GitHub release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GitHubRelease{}, fmt.Errorf("GitHub release lookup failed: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return GitHubRelease{}, fmt.Errorf("decode GitHub release response: %w", err)
	}

	return release, nil
}

// AssetURL searches the release assets for a given name and returns the download URL.
func (gr GitHubRelease) AssetURL(name string) (string, bool) {
	for _, asset := range gr.Assets {
		if asset.Name == name {
			return asset.BrowserDownloadURL, true
		}
	}
	return "", false
}
