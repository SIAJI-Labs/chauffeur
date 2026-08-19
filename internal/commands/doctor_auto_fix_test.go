package commands

import "testing"

func TestDoctorAutoFixCmd_AddsNoconfirmForPacman(t *testing.T) {
	got := doctorAutoFixCmd("sudo pacman -S postgresql-libs")
	want := "sudo pacman -S --noconfirm postgresql-libs"
	if got != want {
		t.Fatalf("doctorAutoFixCmd() = %q, want %q", got, want)
	}
}

func TestDoctorAutoFixCmd_LeavesOtherCommandsUntouched(t *testing.T) {
	cmd := "sudo apt install libpq-dev"
	if got := doctorAutoFixCmd(cmd); got != cmd {
		t.Fatalf("doctorAutoFixCmd() = %q, want %q", got, cmd)
	}
}
