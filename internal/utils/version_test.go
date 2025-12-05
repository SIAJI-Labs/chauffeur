package utils

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

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.2.0", "1.3.0", -1},
		{"1.4.0", "1.3.9", 1},
		{"v1.2.3", "1.2.4", -1},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.current, tt.latest); got != tt.expected {
			t.Fatalf("CompareVersions(%s,%s) = %d, want %d", tt.current, tt.latest, got, tt.expected)
		}
	}
}

func TestGetLatestVersionRespectsDrafts(t *testing.T) {
	setGitHubAPIBase("http://example")
	latestBody, _ := json.Marshal(map[string]any{
		"tag_name":   "v1.2.3",
		"draft":      true,
		"prerelease": false,
	})
	releasesBody, _ := json.Marshal([]map[string]any{
		{"tag_name": "v1.2.2", "draft": false, "prerelease": false},
	})
	restore := clientFactory
	clientFactory = func() *http.Client {
		return &http.Client{
			Transport: &stubTransport{
				responses: map[string]*http.Response{
					"http://example/repos/SIAJI-Labs/chauffeur/releases/latest": {
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(latestBody)),
					},
					"http://example/repos/SIAJI-Labs/chauffeur/releases": {
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(releasesBody)),
					},
				},
			},
		}
	}
	defer func() { clientFactory = restore }()

	latest, err := GetLatestVersion()
	if err != nil {
		t.Fatalf("expected latest version, got error %v", err)
	}
	if latest != "v1.2.2" {
		t.Fatalf("expected v1.2.2 fallback, got %s", latest)
	}
}

func TestIsUpdateAvailable(t *testing.T) {
	setGitHubAPIBase("http://example")
	body, _ := json.Marshal(map[string]any{
		"tag_name":   "v2.0.0",
		"draft":      false,
		"prerelease": false,
	})
	restore := clientFactory
	clientFactory = func() *http.Client {
		return &http.Client{
			Transport: &stubTransport{
				responses: map[string]*http.Response{
					"http://example/repos/SIAJI-Labs/chauffeur/releases/latest": {
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(body)),
					},
				},
			},
		}
	}
	defer func() { clientFactory = restore }()

	ok, latest, err := IsUpdateAvailable("1.0.0")
	if err != nil {
		t.Fatalf("IsUpdateAvailable error: %v", err)
	}
	if !ok || latest != "v2.0.0" {
		t.Fatalf("expected update available to v2.0.0, got ok=%v latest=%s", ok, latest)
	}
}
