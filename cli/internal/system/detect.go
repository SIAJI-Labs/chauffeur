package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Info captures minimal host information needed by installers.
type Info struct {
	Distro string // arch | unknown
	Arch   string // x86_64 | aarch64 | other
	Pretty string
}

/**
 * Detect inspects the local system to gather distro/arch metadata.
 * It prefers Arch detection and falls back to "unknown" when unsure.
 *
 * @return Info describing distro, architecture, and friendly label.
 */
func Detect() (Info, error) {
	distro := detectDistro()
	arch, err := detectArch()
	if err != nil {
		return Info{}, err
	}

	pretty := prettify(distro, arch)

	return Info{
		Distro: distro,
		Arch:   arch,
		Pretty: pretty,
	}, nil
}

/**
 * prettify returns a human-friendly label for the distro/arch combination.
 *
 * @param distro Normalized distro identifier.
 * @param arch   Normalized architecture identifier.
 * @return Combined descriptive string for display.
 */
func prettify(distro, arch string) string {
	label := distro
	if label == "" || label == "unknown" {
		label = "Unknown Linux"
	} else {
		label = strings.ToUpper(label[:1]) + label[1:]
	}
	return fmt.Sprintf("%s (%s)", label, arch)
}

/**
 * detectDistro reads /etc/os-release and attempts to classify the distro.
 *
 * @return Normalized distro identifier, defaulting to "unknown".
 */
func detectDistro() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	var (
		id     string
		idLike []string
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		} else if strings.HasPrefix(line, "ID_LIKE=") {
			raw := strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), `"`)
			idLike = strings.Fields(raw)
		}
	}

	if err := scanner.Err(); err != nil {
		return "unknown"
	}

	if strings.EqualFold(id, "arch") {
		return "arch"
	}

	for _, like := range idLike {
		if strings.EqualFold(like, "arch") {
			return "arch"
		}
	}

	return "unknown"
}

/**
 * detectArch runs uname -m and normalizes the architecture string.
 *
 * @return Normalized architecture or an error when uname fails.
 */
func detectArch() (string, error) {
	cmd := exec.Command("uname", "-m")
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("detect architecture via uname: %w", err)
	}

	arch := strings.TrimSpace(buf.String())
	switch arch {
	case "x86_64":
		return "x86_64", nil
	case "aarch64", "arm64":
		return "aarch64", nil
	default:
		return arch, nil
	}
}
