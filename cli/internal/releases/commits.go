package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GitHubCompareResult captures ahead/behind information between two commits.
type GitHubCompareResult struct {
	Status   string
	AheadBy  int
	BehindBy int
	HeadSHA  string
}

// CompareGitHubCommits fetches comparison metadata between a base commit and a head reference (branch/ref).
func CompareGitHubCommits(client *http.Client, owner, repo, baseSHA, headRef string) (GitHubCompareResult, error) {
	if client == nil {
		client = &http.Client{}
	}

	if baseSHA == "" || headRef == "" {
		return GitHubCompareResult{}, fmt.Errorf("both base commit and head ref are required")
	}

	path := fmt.Sprintf("https://api.github.com/repos/%s/%s/compare/%s...%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(baseSHA),
		url.PathEscape(headRef),
	)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return GitHubCompareResult{}, fmt.Errorf("create compare request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "chauffeur-cli")

	resp, err := client.Do(req)
	if err != nil {
		return GitHubCompareResult{}, fmt.Errorf("compare commits: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GitHubCompareResult{}, fmt.Errorf("compare commits failed: %s", resp.Status)
	}

	var payload struct {
		Status     string `json:"status"`
		AheadBy    int    `json:"ahead_by"`
		BehindBy   int    `json:"behind_by"`
		HeadCommit struct {
			SHA string `json:"sha"`
		} `json:"head_commit"`
		Commits []struct {
			SHA string `json:"sha"`
		} `json:"commits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return GitHubCompareResult{}, fmt.Errorf("decode compare response: %w", err)
	}

	headSHA := payload.HeadCommit.SHA
	if headSHA == "" && len(payload.Commits) > 0 {
		headSHA = payload.Commits[len(payload.Commits)-1].SHA
	}

	return GitHubCompareResult{
		Status:   payload.Status,
		AheadBy:  payload.AheadBy,
		BehindBy: payload.BehindBy,
		HeadSHA:  headSHA,
	}, nil
}

// FetchBranchHeadSHA returns the latest commit SHA for the provided branch/ref.
func FetchBranchHeadSHA(client *http.Client, owner, repo, branch string) (string, error) {
	if client == nil {
		client = &http.Client{}
	}
	if branch == "" {
		return "", fmt.Errorf("branch name is required")
	}

	path := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(branch),
	)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("create branch request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "chauffeur-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch branch head: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("branch lookup failed: %s", resp.Status)
	}

	var payload struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode branch response: %w", err)
	}

	if payload.Commit.SHA == "" {
		return "", fmt.Errorf("branch response missing commit SHA")
	}

	return payload.Commit.SHA, nil
}
