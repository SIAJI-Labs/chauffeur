package helpers

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/siaji/chauffeur/cli/lib"
)

// CaptureOutput captures stdout for the duration of fn.
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	// Save original stdout
	origStdout := os.Stdout
	origCurrentStdout := lib.CurrentStdout
	defer func() {
		os.Stdout = origStdout
		lib.SetOutput(origCurrentStdout)
	}()

	// Create pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	// Redirect both os.Stdout and lib.CurrentStdout
	os.Stdout = w
	lib.SetOutput(w)

	outputCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	// Execute function
	fn()

	// Close write end
	_ = w.Close()

	// Get captured output
	out := <-outputCh
	return out
}
