package lib

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// ProgressReader wraps an io.Reader and renders a live download progress bar.
// On non-TTY outputs it prints a plain "Downloading..." line instead.
type ProgressReader struct {
	r       io.Reader
	total   int64
	label   string
	current int64
	mu      sync.Mutex
}

// NewProgressReader creates a ProgressReader. total is the expected byte count
// (used to draw the progress bar); pass 0 if unknown.
func NewProgressReader(r io.Reader, total int64, label string) *ProgressReader {
	pr := &ProgressReader{r: r, total: total, label: label}
	if !IsTTY {
		fmt.Printf("  ·  %s...\n", label)
	}
	return pr
}

// Read implements io.Reader and updates the progress display on each chunk.
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		pr.mu.Lock()
		pr.current += int64(n)
		current := pr.current
		pr.mu.Unlock()
		pr.render(current)
	}
	return n, err
}

// Done clears the progress line. Call after the read loop completes.
func (pr *ProgressReader) Done() {
	if IsTTY {
		fmt.Printf("\r%-80s\r", "")
	}
}

func (pr *ProgressReader) render(current int64) {
	if !IsTTY {
		return
	}

	var pct int
	if pr.total > 0 {
		pct = int(current * 100 / pr.total)
		if pct > 100 {
			pct = 100
		}
	}

	bar := progressBar(pct, 20)
	totalStr := ""
	if pr.total > 0 {
		totalStr = " / " + FormatBytes(pr.total)
	}

	fmt.Printf("\r  %-14s  %s%s  [%s]  %d%%  ",
		pr.label,
		FormatBytes(current),
		totalStr,
		bar,
		pct,
	)
}

func progressBar(pct, width int) string {
	filled := pct * width / 100
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}
	return bar
}
