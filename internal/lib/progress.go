package lib

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// speedSample records bytes received at a point in time.
type speedSample struct {
	t     time.Time
	bytes int64
}

// ProgressReader wraps an io.Reader and renders a live download progress bar.
// On non-TTY outputs it prints a plain "Downloading..." line instead.
type ProgressReader struct {
	r       io.Reader
	total   int64
	label   string
	current int64
	mu      sync.Mutex

	// Speed tracking: rolling window of the last ~2 seconds of samples.
	samples []speedSample
	started time.Time
}

// NewProgressReader creates a ProgressReader. total is the expected byte count
// (used to draw the progress bar); pass 0 or -1 if unknown.
func NewProgressReader(r io.Reader, total int64, label string) *ProgressReader {
	if total < 0 {
		total = 0
	}
	pr := &ProgressReader{
		r:       r,
		total:   total,
		label:   label,
		started: time.Now(),
	}
	if !IsTTY {
		fmt.Printf("  %s  %s\n", Gray("→"), label)
	}
	return pr
}

// Read implements io.Reader and updates the progress display on each chunk.
func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		now := time.Now()
		pr.mu.Lock()
		pr.current += int64(n)
		current := pr.current
		pr.samples = append(pr.samples, speedSample{t: now, bytes: current})
		// Keep only samples from the last 2 seconds for a rolling average.
		cutoff := now.Add(-2 * time.Second)
		for len(pr.samples) > 1 && pr.samples[0].t.Before(cutoff) {
			pr.samples = pr.samples[1:]
		}
		speed := pr.rollingSpeed()
		pr.mu.Unlock()
		pr.render(current, speed)
	}
	return n, err
}

// rollingSpeed returns bytes/sec based on the current sample window.
// Must be called with pr.mu held.
func (pr *ProgressReader) rollingSpeed() float64 {
	if len(pr.samples) < 2 {
		// Fall back to overall average during the first second.
		elapsed := time.Since(pr.started).Seconds()
		if elapsed < 0.01 {
			return 0
		}
		return float64(pr.current) / elapsed
	}
	first := pr.samples[0]
	last := pr.samples[len(pr.samples)-1]
	dt := last.t.Sub(first.t).Seconds()
	if dt < 0.001 {
		return 0
	}
	return float64(last.bytes-first.bytes) / dt
}

// Done clears the progress line. Call after the read loop completes.
func (pr *ProgressReader) Done() {
	if IsTTY {
		fmt.Printf("\r%-100s\r", "")
	}
}

// spinFrames cycles when total size is unknown.
var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (pr *ProgressReader) render(current int64, speed float64) {
	if !IsTTY {
		return
	}

	speedStr := ""
	if speed > 0 {
		speedStr = "  " + FormatSpeed(speed)
	}

	if pr.total == 0 {
		// Unknown size: show spinner + bytes downloaded + speed.
		frame := spinFrames[(current/65536)%int64(len(spinFrames))]
		fmt.Printf("\r  %-14s  %s  %s%s  ",
			pr.label,
			frame,
			FormatBytes(current),
			speedStr,
		)
		return
	}

	pct := int(current * 100 / pr.total)
	if pct > 100 {
		pct = 100
	}
	fmt.Printf("\r  %-14s  %s / %s  [%s]  %d%%%s  ",
		pr.label,
		FormatBytes(current),
		FormatBytes(pr.total),
		progressBar(pct, 20),
		pct,
		speedStr,
	)
}

// FormatSpeed formats a bytes/sec value as a human-readable string (e.g. "4.2 MB/s").
func FormatSpeed(bps float64) string {
	switch {
	case bps >= 1<<30:
		return fmt.Sprintf("%.1f GB/s", bps/(1<<30))
	case bps >= 1<<20:
		return fmt.Sprintf("%.1f MB/s", bps/(1<<20))
	case bps >= 1<<10:
		return fmt.Sprintf("%.0f KB/s", bps/(1<<10))
	default:
		return fmt.Sprintf("%.0f B/s", bps)
	}
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
