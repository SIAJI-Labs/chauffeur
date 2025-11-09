package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// CaptureOutput captures stdout for the duration of fn.
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	outputCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = origStdout
	out := <-outputCh
	return out
}
