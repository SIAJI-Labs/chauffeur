package services

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// IsPortAvailable returns true if the given TCP port can be bound.
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindProcessOnPort returns the PID and process name using the given port,
// by scanning /proc/net/tcp and /proc/net/tcp6.
// Returns (0, "", nil) if nothing is found.
func FindProcessOnPort(port int) (pid int, name string, err error) {
	hexPort := fmt.Sprintf("%04X", port)

	for _, netFile := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		data, err := os.ReadFile(netFile)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines[1:] { // skip header
			fields := strings.Fields(line)
			if len(fields) < 10 {
				continue
			}
			// local_address field: "0100007F:1F90" (IP:PORT in hex)
			localAddr := fields[1]
			parts := strings.Split(localAddr, ":")
			if len(parts) != 2 {
				continue
			}
			if !strings.EqualFold(parts[1], hexPort) {
				continue
			}
			// inode is field 9
			inode, err := strconv.Atoi(fields[9])
			if err != nil {
				continue
			}
			pid, name = findPIDByInode(inode)
			if pid > 0 {
				return pid, name, nil
			}
		}
	}
	return 0, "", nil
}

// findPIDByInode scans /proc/<pid>/fd/ looking for a socket with the given inode.
func findPIDByInode(inode int) (int, string) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, ""
	}
	target := fmt.Sprintf("socket:[%d]", inode)
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // not a PID dir
		}
		fdDir := fmt.Sprintf("/proc/%d/fd", pid)
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				name := processName(pid)
				return pid, name
			}
		}
	}
	return 0, ""
}

// processName reads the process name from /proc/<pid>/comm.
func processName(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}
