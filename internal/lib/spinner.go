package lib

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Spinner displays an animated spinner while work is in progress.
// On non-TTY outputs it prints a plain status line instead.
type Spinner struct {
	label string
	done  chan struct{}
	once  sync.Once
}

// NewSpinner creates and starts a spinner with the given message.
func NewSpinner(message string) *Spinner {
	s := &Spinner{
		label: message,
		done:  make(chan struct{}),
	}
	if IsTTY {
		go s.spin()
	} else {
		fmt.Printf("  %s  %s\n", Gray("→"), message)
	}
	return s
}

func (s *Spinner) spin() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	t := time.NewTicker(80 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-t.C:
			fmt.Printf("\r  %s  %s  ", Cyan(frames[i%len(frames)]), s.label)
			i++
		}
	}
}

func (s *Spinner) stop() {
	s.once.Do(func() { close(s.done) })
	time.Sleep(20 * time.Millisecond) // let goroutine flush
}

func (s *Spinner) clear() {
	if IsTTY {
		fmt.Printf("\r%-80s\r", "")
	}
}

// Success stops the spinner and prints a green success line.
// detail is printed in gray beneath the label (pass "" to omit).
func (s *Spinner) Success(detail string) {
	s.stop()
	s.clear()
	fmt.Printf("  %s  %s\n", Green("✓"), s.label)
	if detail != "" {
		fmt.Printf("  %-14s  %s\n", "", Gray(detail))
	}
}

// Fail stops the spinner and prints a red failure line to stderr.
// detail is printed in red beneath the label (pass "" to omit).
func (s *Spinner) Fail(detail string) {
	s.stop()
	s.clear()
	fmt.Fprintf(os.Stderr, "  %s  %s\n", Red("✗"), s.label)
	if detail != "" {
		fmt.Fprintf(os.Stderr, "  %-14s  %s\n", "", Red(detail))
	}
}
