package commands

import "testing"

func TestNormalizeLinkArgsSupportsPathBeforeFlags(t *testing.T) {
	got := normalizeLinkArgs([]string{"/tmp/project", "--php", "8.3", "--no-interactive"})
	want := []string{"--php", "8.3", "--no-interactive", "/tmp/project"}
	if len(got) != len(want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args = %#v, want %#v", got, want)
		}
	}
}

func TestNormalizeLinkArgsPreservesFlagsFirstForm(t *testing.T) {
	input := []string{"--php", "8.3", "/tmp/project"}
	got := normalizeLinkArgs(input)
	if len(got) != len(input) {
		t.Fatalf("args = %#v, want %#v", got, input)
	}
	for i := range input {
		if got[i] != input[i] {
			t.Fatalf("args = %#v, want %#v", got, input)
		}
	}
}
