package runtime

import (
	"context"
	"testing"
)

func TestPodmanLogsBuildsScopedArguments(t *testing.T) {
	runner := &recordingRunner{result: CommandResult{Stdout: "ready\n"}}
	logs, err := (Podman{Runner: runner}).Logs(context.Background(), Scope{Version: "8.3"}, LogOptions{Lines: 25})
	if err != nil || logs != "ready\n" {
		t.Fatalf("logs = %q, err = %v", logs, err)
	}
	want := []string{"logs", "--tail", "25", "chauf-php83-fpm"}
	for i := range want {
		if runner.args[0][i] != want[i] {
			t.Fatalf("args = %#v, want %#v", runner.args, want)
		}
	}
}
