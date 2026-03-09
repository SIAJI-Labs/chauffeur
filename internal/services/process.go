package services

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ── PID file helpers ──────────────────────────────────────────────────────────

func readPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in %s: %w", path, err)
	}
	return pid, nil
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// waitForPID polls until the PID file appears and contains a running process.
func waitForPID(pidPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pid, err := readPIDFile(pidPath); err == nil && processAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("process did not start within %s", timeout)
}

// waitForExit polls until the process is gone or timeout is exceeded,
// then sends SIGKILL if still alive.
func waitForExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	// Graceful timeout exceeded — force kill
	_ = syscall.Kill(pid, syscall.SIGKILL)
	return nil
}

// waitForSocket polls until the socket file appears.
func waitForSocket(sockPath string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sockPath); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("socket not ready after %s: %s", timeout, sockPath)
}

// ── Process metrics ───────────────────────────────────────────────────────────

// processUptime reads /proc/<pid>/stat to compute how long the process has been running.
func processUptime(pid int) time.Duration {
	// Field 22 in /proc/<pid>/stat is the start time in clock ticks since boot.
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0
	}
	// Find the closing ) of the comm field and skip past it.
	s := string(data)
	idx := strings.LastIndex(s, ")")
	if idx < 0 {
		return 0
	}
	fields := strings.Fields(s[idx+1:])
	if len(fields) < 20 {
		return 0
	}
	startTicks, err := strconv.ParseInt(fields[19], 10, 64) // field 22 = index 19 after )
	if err != nil {
		return 0
	}

	// Boot time from /proc/uptime (seconds since boot, with uptime of pid relative to it)
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	parts := strings.Fields(string(uptimeData))
	if len(parts) < 1 {
		return 0
	}
	systemUptimeSec, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	const clockTicksPerSec = 100 // USER_HZ on Linux
	processStartSec := float64(startTicks) / clockTicksPerSec
	processUptimeSec := systemUptimeSec - processStartSec
	if processUptimeSec < 0 {
		return 0
	}
	return time.Duration(processUptimeSec * float64(time.Second))
}

// processMemoryMB reads VmRSS from /proc/<pid>/status and returns MB.
func processMemoryMB(pid int) int {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.Atoi(fields[1])
				if err == nil {
					return kb / 1024
				}
			}
		}
	}
	return 0
}

// FormatUptime formats a duration as "2h 34m" or "45s".
func FormatUptime(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
