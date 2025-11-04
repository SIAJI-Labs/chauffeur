package lib

import (
	"fmt"
	"strings"
	"time"

	"github.com/siaji/chauffeur/cli/internal/logging"
)

// progressPrinter provides real-time progress tracking for downloads and file operations
type progressPrinter struct {
	logger     *logging.CommandLogger
	label      string
	total      int64
	current    int64
	width      int
	lastPrint  time.Time
	done       bool
}

// NewProgressPrinter creates a new progress printer with command logger integration
func NewProgressPrinter(label string, total int64) *progressPrinter {
	logger := logging.NewCommandLogger("install")
	return &progressPrinter{
		logger:    logger,
		label:     label,
		total:     total,
		width:     32,
		lastPrint: time.Time{},
	}
}

// NewProgressPrinterWithLogger creates a new progress printer with a specific logger
func NewProgressPrinterWithLogger(label string, total int64, logger *logging.CommandLogger) *progressPrinter {
	return &progressPrinter{
		logger:    logger,
		label:     label,
		total:     total,
		width:     32,
		lastPrint: time.Time{},
	}
}

// Write implements io.Writer to track progress as data is written
func (p *progressPrinter) Write(b []byte) (int, error) {
	n := len(b)
	p.current += int64(n)
	if time.Since(p.lastPrint) >= 120*time.Millisecond || p.current == p.total {
		p.render()
	}
	return n, nil
}

// Finish completes the progress display and shows final status
func (p *progressPrinter) Finish() {
	if p == nil || p.done {
		return
	}
	p.renderFinal()
	p.done = true
}

// render updates the progress display on the current line
func (p *progressPrinter) render() {
	if p.total > 0 {
		ratio := float64(p.current) / float64(p.total)
		if ratio > 1 {
			ratio = 1
		}
		filled := int(ratio * float64(p.width))
		if filled > p.width {
			filled = p.width
		}
		bar := strings.Repeat("#", filled) + strings.Repeat(".", p.width-filled)
		// Use command logger prefix format and clear line
		fmt.Printf("\r\033[K%s %s... [%s] %3.0f%%", p.logger.Prefix(), p.label, bar, ratio*100)
	} else {
		fmt.Printf("\r\033[K%s %s... %s", p.logger.Prefix(), p.label, HumanBytes(p.current))
	}
	// Ensure the output is flushed immediately
	fmt.Print("")
	p.lastPrint = time.Now()
}

// renderFinal displays the final completed progress with newline
func (p *progressPrinter) renderFinal() {
	if p.total > 0 {
		fmt.Printf("\r\033[K%s %s... [%s] 100%% (%s)\n", p.logger.Prefix(), p.label, strings.Repeat("#", p.width), HumanBytes(p.current))
	} else {
		fmt.Printf("\r\033[K%s %s... %s\n", p.logger.Prefix(), p.label, HumanBytes(p.current))
	}
}

// HumanBytes formats byte counts in human-readable format
func HumanBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	value := float64(n)
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	if idx == 0 {
		return fmt.Sprintf("%d %s", n, units[idx])
	}
	return fmt.Sprintf("%.1f %s", value, units[idx])
}
