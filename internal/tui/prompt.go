package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/siegg/chauffeur/internal/lib"
)

// Confirm reads a conservative yes/no answer. It never waits in a non-TTY.
func Confirm(prompt string) bool {
	if !Interactive() {
		return false
	}
	fmt.Printf("  %s %s: ", prompt, lib.Gray("[y/N]"))
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(answer) == "" {
		return false
	}
	answer = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(answer, "\r", "")))
	return answer == "y" || answer == "yes"
}

// ConfirmText reads an exact confirmation phrase for destructive operations.
func ConfirmText(prompt, expected string) bool {
	if !Interactive() {
		return false
	}
	fmt.Printf("  %s: ", prompt)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(answer) == "" {
		return false
	}
	return strings.TrimSpace(answer) == expected
}
