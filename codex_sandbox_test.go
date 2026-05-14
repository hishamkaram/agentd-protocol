package protocol

import "testing"

func TestCodexSandboxModeConstants(t *testing.T) {
	t.Parallel()

	tests := map[string]CodexSandboxMode{
		"read-only":          CodexSandboxReadOnly,
		"workspace-write":    CodexSandboxWorkspaceWrite,
		"danger-full-access": CodexSandboxDangerFullAccess,
	}
	for want, got := range tests {
		if string(got) != want {
			t.Errorf("CodexSandboxMode constant = %q, want %q", got, want)
		}
	}
}

func TestKnownCodexSandboxModes(t *testing.T) {
	t.Parallel()

	got := KnownCodexSandboxModes()
	want := []CodexSandboxMode{
		CodexSandboxReadOnly,
		CodexSandboxWorkspaceWrite,
		CodexSandboxDangerFullAccess,
	}
	if len(got) != len(want) {
		t.Fatalf("KnownCodexSandboxModes len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("KnownCodexSandboxModes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	got[0] = CodexSandboxDangerFullAccess
	if KnownCodexSandboxModes()[0] != CodexSandboxReadOnly {
		t.Error("KnownCodexSandboxModes must return a fresh slice")
	}
}

func TestIsKnownCodexSandboxMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode CodexSandboxMode
		want bool
	}{
		{name: "read-only", mode: CodexSandboxReadOnly, want: true},
		{name: "workspace-write", mode: CodexSandboxWorkspaceWrite, want: true},
		{name: "danger-full-access", mode: CodexSandboxDangerFullAccess, want: true},
		{name: "empty", mode: "", want: false},
		{name: "unknown", mode: "sandbox-but-not-really", want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsKnownCodexSandboxMode(tt.mode); got != tt.want {
				t.Errorf("IsKnownCodexSandboxMode(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}
