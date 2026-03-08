package installers_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/siegg/chauffeur/internal/installers"
)

// ── VerifySHA256 ──────────────────────────────────────────────────────────────

func TestVerifySHA256_Match(t *testing.T) {
	f := writeTempFile(t, "hello chauffeur\n")
	hash := sha256Hex(t, "hello chauffeur\n")
	if err := installers.VerifySHA256(f, hash); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySHA256_Mismatch(t *testing.T) {
	f := writeTempFile(t, "hello chauffeur\n")
	if err := installers.VerifySHA256(f, "deadbeef"); err == nil {
		t.Fatal("expected error for wrong hash, got nil")
	}
}

func TestVerifySHA256_MissingFile(t *testing.T) {
	if err := installers.VerifySHA256(filepath.Join(t.TempDir(), "no-such-file"), "abc"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// ── ExtractTarGz ─────────────────────────────────────────────────────────────

func TestExtractTarGz_Basic(t *testing.T) {
	tarPath := makeTarGz(t, "mylib-1.0", map[string]string{
		"mylib-1.0/README": "hello",
		"mylib-1.0/src/main.c": "int main(){}",
	})
	dest := t.TempDir()

	topDir, err := installers.ExtractTarGz(tarPath, dest, 0)
	if err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}
	if topDir != filepath.Join(dest, "mylib-1.0") {
		t.Errorf("topDir = %q, want %q", topDir, filepath.Join(dest, "mylib-1.0"))
	}

	assertFileContent(t, filepath.Join(dest, "mylib-1.0", "README"), "hello")
	assertFileContent(t, filepath.Join(dest, "mylib-1.0", "src", "main.c"), "int main(){}")
}

func TestExtractTarGz_Strip(t *testing.T) {
	tarPath := makeTarGz(t, "mylib-1.0", map[string]string{
		"mylib-1.0/file.txt": "stripped",
	})
	dest := t.TempDir()

	if _, err := installers.ExtractTarGz(tarPath, dest, 1); err != nil {
		t.Fatalf("ExtractTarGz strip=1: %v", err)
	}
	assertFileContent(t, filepath.Join(dest, "file.txt"), "stripped")
}

func TestExtractTarGz_PathTraversal(t *testing.T) {
	// Build a tarball with a path-traversal entry manually.
	tarPath := makeTarGzRaw(t, []tarEntry{
		{Name: "../evil.txt", Content: "pwned", Mode: 0644},
	})
	dest := t.TempDir()

	if _, err := installers.ExtractTarGz(tarPath, dest, 0); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	evilPath := filepath.Join(filepath.Dir(dest), "evil.txt")
	if _, err := os.Stat(evilPath); err == nil {
		t.Error("path-traversal file was written outside dest directory")
	}
}

// ── MajorMinor ────────────────────────────────────────────────────────────────

func TestMajorMinor(t *testing.T) {
	cases := []struct{ in, want string }{
		{"8.3", "8.3"},
		{"8.3.14", "8.3"},
		{"7.4.33", "7.4"},
		{"php8.3", "8.3"},
		{"php7.4.33", "7.4"},
		{"8", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := installers.MajorMinor(tc.in); got != tc.want {
			t.Errorf("MajorMinor(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── DownloadWithProgress ──────────────────────────────────────────────────────

func TestDownloadWithProgress_Cached(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "file.tar.gz")
	os.WriteFile(dest, []byte("existing"), 0644)

	cached, err := installers.DownloadWithProgress("http://should-not-be-called", dest, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cached {
		t.Error("expected cached=true for existing file")
	}
}

func TestDownloadWithProgress_Success(t *testing.T) {
	body := strings.Repeat("x", 4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "downloaded.bin")
	cached, err := installers.DownloadWithProgress(srv.URL+"/file", dest, false)
	if err != nil {
		t.Fatalf("DownloadWithProgress: %v", err)
	}
	if cached {
		t.Error("expected cached=false for fresh download")
	}
	got, _ := os.ReadFile(dest)
	if string(got) != body {
		t.Errorf("downloaded content mismatch: got %d bytes, want %d", len(got), len(body))
	}
}

func TestDownloadWithProgress_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.bin")
	if _, err := installers.DownloadWithProgress(srv.URL+"/missing", dest, false); err == nil {
		t.Fatal("expected error for HTTP 404, got nil")
	}
	if _, err := os.Stat(dest); err == nil {
		t.Error("temp file should not persist after failed download")
	}
}

func TestDownloadWithProgress_NoCache(t *testing.T) {
	body := "fresh content"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "file.bin")
	// Write stale content.
	os.WriteFile(dest, []byte("stale"), 0644)

	cached, err := installers.DownloadWithProgress(srv.URL+"/file", dest, true)
	if err != nil {
		t.Fatalf("DownloadWithProgress noCache: %v", err)
	}
	if cached {
		t.Error("expected cached=false when noCache=true")
	}
	got, _ := os.ReadFile(dest)
	if string(got) != body {
		t.Errorf("got %q, want %q", got, body)
	}
}

// ── FetchChecksum ─────────────────────────────────────────────────────────────

func TestFetchChecksum(t *testing.T) {
	checksumFile := "abc123  php-8.3.14.tar.gz\ndef456  php-8.2.27.tar.gz\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, checksumFile)
	}))
	defer srv.Close()

	got, err := installers.FetchChecksum(srv.URL+"/checksums", "php-8.3.14.tar.gz")
	if err != nil {
		t.Fatalf("FetchChecksum: %v", err)
	}
	if got != "abc123" {
		t.Errorf("FetchChecksum = %q, want %q", got, "abc123")
	}
}

func TestFetchChecksum_StarPrefix(t *testing.T) {
	// sha256sum uses "hash *filename" format.
	checksumFile := "abc123 *php-8.3.14.tar.gz\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, checksumFile)
	}))
	defer srv.Close()

	got, err := installers.FetchChecksum(srv.URL+"/checksums", "php-8.3.14.tar.gz")
	if err != nil {
		t.Fatalf("FetchChecksum: %v", err)
	}
	if got != "abc123" {
		t.Errorf("FetchChecksum = %q, want %q", got, "abc123")
	}
}

func TestFetchChecksum_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "abc123  other-file.tar.gz\n")
	}))
	defer srv.Close()

	if _, err := installers.FetchChecksum(srv.URL+"/checksums", "missing.tar.gz"); err == nil {
		t.Fatal("expected error for missing filename, got nil")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprint(f, content)
	f.Close()
	return f.Name()
}

func sha256Hex(t *testing.T, content string) string {
	t.Helper()
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("file %s content = %q, want %q", path, got, want)
	}
}

type tarEntry struct {
	Name    string
	Content string
	Mode    int64
	IsDir   bool
}

// makeTarGz creates a .tar.gz with a top-level directory named topDir
// and files under it, returning the archive path.
func makeTarGz(t *testing.T, topDir string, files map[string]string) string {
	t.Helper()
	entries := []tarEntry{{Name: topDir + "/", IsDir: true, Mode: 0755}}
	for name, content := range files {
		entries = append(entries, tarEntry{Name: name, Content: content, Mode: 0644})
	}
	return makeTarGzRaw(t, entries)
}

func makeTarGzRaw(t *testing.T, entries []tarEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for _, e := range entries {
		hdr := &tar.Header{
			Name: e.Name,
			Mode: e.Mode,
			Size: int64(len(e.Content)),
		}
		if e.IsDir {
			hdr.Typeflag = tar.TypeDir
			hdr.Size = 0
		} else {
			hdr.Typeflag = tar.TypeReg
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if !e.IsDir {
			if _, err := tw.Write([]byte(e.Content)); err != nil {
				t.Fatal(err)
			}
		}
	}

	tw.Close()
	gw.Close()
	f.Close()
	return path
}
