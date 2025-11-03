package tests

import (
	"bytes"
	"io"
	"os"
)

type execFunc func() error

type outputResult struct {
	output string
	err    error
}

func captureOutput(fn execFunc) string {
	res := capture(fn)
	if res.err != nil {
		panic(res.err)
	}
	return res.output
}

func captureError(fn execFunc) string {
	res := capture(fn)
	if res.err != nil {
		return res.err.Error()
	}
	return res.output
}

func capture(fn execFunc) outputResult {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wOut

	err := fn()

	_ = wOut.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	_ = rOut.Close()

	return outputResult{output: buf.String(), err: err}
}
