package installers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// logToolFailure writes a structured error entry into the tool log directory.
func logToolFailure(prefix, component, action, version string, err error) (string, error) {
	logDir := filepath.Join(prefix, "logs", component)
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		return "", fmt.Errorf("ensure tool log dir: %w", mkErr)
	}

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	nameParts := []string{sanitizePathToken(action)}
	if version != "" {
		nameParts = append(nameParts, sanitizePathToken(version))
	}
	filename := fmt.Sprintf("%s-%s.log", strings.Join(nameParts, "-"), now.Format("20060102T150405Z"))
	logPath := filepath.Join(logDir, filename)

	file, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if openErr != nil {
		return "", fmt.Errorf("open tool log: %w", openErr)
	}
	defer file.Close()

	summary := strings.TrimSpace(err.Error())
	if summary == "" {
		summary = "<no error summary>"
	}

	detail := summary
	var detailProvider detailedError
	if errors.As(err, &detailProvider) {
		if d := strings.TrimSpace(detailProvider.Detail()); d != "" {
			detail = d
		}
	}
	if detail == "" {
		detail = "<no error detail>"
	}
	indentedDetail := "    " + strings.ReplaceAll(detail, "\n", "\n    ")

	entry := fmt.Sprintf("[%s] ERROR %s %s %s\nSummary: %s\nDetails:\n%s\n\n",
		timestamp, component, action, version, summary, indentedDetail)
	if _, writeErr := file.WriteString(entry); writeErr != nil {
		return "", fmt.Errorf("write tool log: %w", writeErr)
	}

	return logPath, nil
}

func sanitizePathToken(token string) string {
	if token == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
