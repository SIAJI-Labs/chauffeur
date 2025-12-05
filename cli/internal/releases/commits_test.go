package releases

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type stubTransport struct {
	responses map[string]*http.Response
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if resp, ok := s.responses[req.URL.String()]; ok {
		return resp, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewBufferString("not found")),
	}, nil
}

func TestCompareGitHubCommits(t *testing.T) {
	setGitHubAPIBase("http://example")
	body, _ := json.Marshal(map[string]any{
		"status":    "ahead",
		"ahead_by":  2,
		"behind_by": 0,
		"head_commit": map[string]any{
			"sha": "abc123",
		},
		"commits": []map[string]any{
			{"sha": "abc123"},
		},
	})
	client := &http.Client{
		Transport: &stubTransport{
			responses: map[string]*http.Response{
				"http://example/repos/me/repo/compare/base...head": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
				},
			},
		},
	}

	result, err := CompareGitHubCommits(client, "me", "repo", "base", "head")
	if err != nil {
		t.Fatalf("compare should succeed: %v", err)
	}
	if result.Status != "ahead" || result.AheadBy != 2 || result.HeadSHA != "abc123" {
		t.Fatalf("unexpected compare result: %+v", result)
	}
}

func TestFetchBranchHeadSHA(t *testing.T) {
	setGitHubAPIBase("http://example")
	body, _ := json.Marshal(map[string]any{
		"commit": map[string]any{"sha": "deadbeef"},
	})
	client := &http.Client{
		Transport: &stubTransport{
			responses: map[string]*http.Response{
				"http://example/repos/me/repo/branches/main": {
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(body)),
				},
			},
		},
	}

	sha, err := FetchBranchHeadSHA(client, "me", "repo", "main")
	if err != nil {
		t.Fatalf("fetch head should succeed: %v", err)
	}
	if sha != "deadbeef" {
		t.Fatalf("expected sha deadbeef, got %s", sha)
	}
}
