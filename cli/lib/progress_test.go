package lib

import (
	"bytes"
	"testing"
	"time"
)

func TestProgressPrinterRenderFinalBytes(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	p := NewProgressPrinterWithLogger("download file.tar.gz", 0, NewCommandLogger("install"))
	p.Write([]byte("hello"))
	p.Finish()

	if !bytes.Contains(buf.Bytes(), []byte("file.tar.gz")) {
		t.Fatalf("expected output to mention label, got %s", buf.String())
	}
}

func TestProgressPrinterRenderBar(t *testing.T) {
	restore := CurrentStdout
	defer func() { CurrentStdout = restore }()

	buf := &bytes.Buffer{}
	CurrentStdout = buf

	p := NewProgressPrinterWithLogger("download file.tar.gz", 10, NewCommandLogger("install"))
	p.Write([]byte("12345"))
	time.Sleep(10 * time.Millisecond) // allow throttling to pass
	p.Write([]byte("12345"))
	p.Finish()

	if !bytes.Contains(buf.Bytes(), []byte("100%")) {
		t.Fatalf("expected progress to reach 100%%, got %s", buf.String())
	}
}
