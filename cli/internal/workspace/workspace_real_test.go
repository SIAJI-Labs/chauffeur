package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

type stubLogger struct {
	block   []string
	info    []string
	prompts []string
	success []string
}

func (s *stubLogger) PrintBlock(msg string)   { s.block = append(s.block, msg) }
func (s *stubLogger) Info(msg string)         { s.info = append(s.info, msg) }
func (s *stubLogger) Prompt(msg, _ string)    { s.prompts = append(s.prompts, msg) }
func (s *stubLogger) Success(msg, ctx string) { s.success = append(s.success, msg+" "+ctx) }

func TestDirAndEnsure(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	root, err := Dir()
	if err != nil {
		t.Fatalf("Dir error: %v", err)
	}
	if err := Ensure("config", "projects"); err != nil {
		t.Fatalf("Ensure error: %v", err)
	}
	for _, p := range []string{root, filepath.Join(root, "config"), filepath.Join(root, "projects")} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected path to exist: %s", p)
		}
	}
}

func TestValidateOrPromptWithLoggerInitializesWorkspace(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	logger := &stubLogger{}
	ok, err := ValidateOrPromptWithLogger(logger)
	if err != nil {
		t.Fatalf("validate prompt error: %v", err)
	}
	if !ok {
		t.Fatalf("expected workspace to be ready")
	}
	if len(logger.prompts) == 0 {
		t.Fatalf("expected prompt to be issued")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".chauffeur")); err != nil {
		t.Fatalf("workspace not created: %v", err)
	}
}

func TestValidateForCommandWithHelpSkipsInit(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	ok, err := ValidateForCommandWithLogger([]string{"--help"}, &stubLogger{})
	if err != nil {
		t.Fatalf("ValidateForCommandWithLogger error: %v", err)
	}
	if !ok {
		t.Fatalf("expected help flag to bypass validation")
	}
}
