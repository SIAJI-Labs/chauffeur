package panel

import (
	"bytes"
	"log"
	"sync"
	"strings"
	"testing"
)

var loggerMu sync.Mutex

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()

	loggerMu.Lock()
	t.Cleanup(loggerMu.Unlock)

	var buf bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})

	fn()
	return buf.String()
}

func TestLogListDatabasesRequestSuccess(t *testing.T) {
	output := captureLogs(t, func() {
		logListDatabasesResult("chauf-mysql8", "mysql8", []string{"app", "analytics"}, nil)
	})

	if !strings.Contains(output, `panel: list databases container="chauf-mysql8" engine="mysql8" count=2`) {
		t.Fatalf("expected success log, got %q", output)
	}

	if strings.Contains(output, "error=") {
		t.Fatalf("did not expect error in success log, got %q", output)
	}
}

func TestLogListDatabasesRequestFailure(t *testing.T) {
	output := captureLogs(t, func() {
		logListDatabasesResult("chauf-postgres", "postgres", nil, assertErr("list failed"))
	})

	if !strings.Contains(output, `panel: list databases container="chauf-postgres" engine="postgres" count=0 error="list failed"`) {
		t.Fatalf("expected failure log, got %q", output)
	}
}

func TestLogCreateBackupRequestLifecycle(t *testing.T) {
	output := captureLogs(t, func() {
		logCreateBackupRequest("chauf-mysql8", []DatabaseBackup{{Name: "app"}, {Name: "analytics"}})
		logCreateBackupStart("chauf-mysql8", "app")
		logCreateBackupSuccess("chauf-mysql8", "app", "chauf-mysql8-app-20260604-120000.tar.gz")
		logCreateBackupFailure("chauf-mysql8", "analytics", assertErr("backup failed"))
		logCreateBackupSummary("chauf-mysql8", 1, 1)
	})

	expected := []string{
		`panel: create backup container="chauf-mysql8" requested_databases=[app analytics]`,
		`panel: backup start container="chauf-mysql8" database="app"`,
		`panel: backup success container="chauf-mysql8" database="app" filename="chauf-mysql8-app-20260604-120000.tar.gz"`,
		`panel: backup failure container="chauf-mysql8" database="analytics" error="backup failed"`,
		`panel: create backup summary container="chauf-mysql8" succeeded=1 failed=1`,
	}

	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Fatalf("expected log line %q in output %q", want, output)
		}
	}
}

func TestLogListDatabasesRequestFailureNilErrorDoesNotPanic(t *testing.T) {
	output := captureLogs(t, func() {
		logListDatabasesResult("chauf-postgres", "postgres", nil, nil)
	})

	if !strings.Contains(output, `panel: list databases container="chauf-postgres" engine="postgres" count=0`) {
		t.Fatalf("expected nil-error log, got %q", output)
	}

	if strings.Contains(output, "error=") {
		t.Fatalf("did not expect error field for nil error, got %q", output)
	}
}

func TestLogCreateBackupFailureNilErrorDoesNotPanic(t *testing.T) {
	output := captureLogs(t, func() {
		logCreateBackupFailure("chauf-mysql8", "analytics", nil)
	})

	if !strings.Contains(output, `panel: backup failure container="chauf-mysql8" database="analytics" error="<nil>"`) {
		t.Fatalf("expected nil-error failure log, got %q", output)
	}
}

type testError string

func (e testError) Error() string { return string(e) }

func assertErr(message string) error { return testError(message) }
