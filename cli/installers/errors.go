package installers

import (
	"strings"
)

// detailedError exposes rich diagnostic output for logging purposes while keeping CLI output concise.
type detailedError interface {
	error
	Detail() string
}

type commandError struct {
	Name   string
	Args   []string
	Err    error
	Stdout string
	Stderr string
}

func (c commandError) Error() string {
	cmd := strings.TrimSpace(strings.Join(append([]string{c.Name}, c.Args...), " "))
	if cmd == "" {
		cmd = c.Name
	}
	if c.Err != nil {
		return strings.TrimSpace(cmd + " failed: " + c.Err.Error())
	}
	return cmd + " failed"
}

func (c commandError) Detail() string {
	cmd := strings.TrimSpace(strings.Join(append([]string{c.Name}, c.Args...), " "))
	builder := strings.Builder{}
	builder.WriteString(cmd + " failed")
	if c.Err != nil {
		builder.WriteString(": " + c.Err.Error())
	}
	builder.WriteString("\nstdout:\n")
	if c.Stdout != "" {
		builder.WriteString(c.Stdout)
	} else {
		builder.WriteString("<empty>\n")
	}
	builder.WriteString("\nstderr:\n")
	if c.Stderr != "" {
		builder.WriteString(c.Stderr)
	} else {
		builder.WriteString("<empty>\n")
	}
	return builder.String()
}
