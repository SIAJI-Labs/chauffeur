package runtime

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPodmanPHP85ParityFixture(t *testing.T) {
	runPHPParityFixture(t, "8.5", 18085)
}

func TestPodmanPHP80ParityFixture(t *testing.T) {
	runPHPParityFixture(t, "8.0", 18080)
}

func TestPodmanPHP74ParityFixture(t *testing.T) {
	runPHPParityFixture(t, "7.4", 18074)
}

func runPHPParityFixture(t *testing.T, version string, port int) {
	if os.Getenv("CHAUFFEUR_PODMAN_INTEGRATION") != "1" {
		t.Skip("set CHAUFFEUR_PODMAN_INTEGRATION=1 to run Podman integration tests")
	}
	runner := ExecRunner{}
	for _, image := range []string{PHPImage(version), NginxImage} {
		result, err := runner.Run(context.Background(), "image", "exists", image)
		if err != nil || result.ExitCode != 0 {
			t.Skipf("required image unavailable: %s", image)
		}
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.php"), []byte("<?php echo 'php"+version+"-ok';"), 0644); err != nil {
		t.Fatal(err)
	}
	certDir := filepath.Join(root, "certs")
	if err := writeFixtureCertificate(certDir); err != nil {
		t.Fatal(err)
	}
	config, err := RenderNginxPHPConfigForRoutesWithHTTPS([]NginxRoute{{ServerName: "fixture.test", DocumentRoot: "/workspace", Upstream: FPMContainerName(version), SSL: true, CertName: "fixture.test"}}, 8080, 8443)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "nginx.conf")
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	rt := Podman{Runner: runner}
	ctx := context.Background()
	if err := rt.EnsurePHPContainer(ctx, Scope{Version: version, Project: root}, PHPImage(version), root); err != nil {
		t.Fatal(err)
	}
	for _, command := range [][]string{{"php", "-v"}, {"composer", "--version"}} {
		if _, err := rt.Exec(ctx, Scope{Version: version}, command, ExecOptions{}); err != nil {
			t.Fatalf("%v: %v", command, err)
		}
	}
	if err := rt.EnsureNginxContainerWithRootsAndPorts(ctx, configPath, map[string]string{"/workspace": root}, port, port+1000, certDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = runner.Run(ctx, "container", "rm", "-f", "chauf-nginx")
		_, _ = runner.Run(ctx, "container", "rm", "-f", FPMContainerName(version))
	})
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	var response *http.Response
	for attempt := 0; attempt < 20; attempt++ {
		request, requestErr := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/", port), nil)
		if requestErr == nil {
			request.Host = "fixture.test"
			response, err = client.Do(request)
		}
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusMovedPermanently || string(body) == "php"+version+"-ok" {
		t.Fatalf("HTTP status=%d body=%q err=%v", response.StatusCode, body, err)
	}
	secureClient := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} // fixture certificate is intentionally self-signed
	secureRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d/", port+1000), nil)
	if err != nil {
		t.Fatal(err)
	}
	secureRequest.Host = "fixture.test"
	secureResponse, err := secureClient.Do(secureRequest)
	if err != nil {
		t.Fatal(err)
	}
	defer secureResponse.Body.Close()
	secureBody, err := io.ReadAll(secureResponse.Body)
	if err != nil || string(secureBody) != "php"+version+"-ok" {
		t.Fatalf("HTTPS status=%d body=%q err=%v", secureResponse.StatusCode, secureBody, err)
	}
}

func writeFixtureCertificate(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialBits))
	if err != nil {
		return err
	}
	template := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: "fixture.test"}, DNSNames: []string{"fixture.test"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return err
	}
	certFile, err := os.Create(filepath.Join(dir, "fixture.test.crt"))
	if err != nil {
		return err
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		_ = certFile.Close()
		return err
	}
	if err := certFile.Close(); err != nil {
		return err
	}
	keyFile, err := os.Create(filepath.Join(dir, "fixture.test.key"))
	if err != nil {
		return err
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		_ = keyFile.Close()
		return err
	}
	return keyFile.Close()
}

const serialBits = 128
