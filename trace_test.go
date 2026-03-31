package protocol_test

import (
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestNewTraceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "generates 32 char hex string"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tid := protocol.NewTraceID()
			if len(tid) != 32 {
				t.Errorf("NewTraceID() length = %d, want 32", len(tid))
			}
			if !protocol.ValidTraceID(tid) {
				t.Errorf("NewTraceID() produced invalid trace ID: %q", tid)
			}
		})
	}
}

func TestNewTraceID_Uniqueness(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		tid := protocol.NewTraceID()
		if _, exists := seen[tid]; exists {
			t.Fatalf("duplicate trace ID after %d iterations: %q", i, tid)
		}
		seen[tid] = struct{}{}
	}
}

func TestValidTraceID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid 32 hex", input: "4bf92f3577b34da6a3ce929d0e0e4736", want: true},
		{name: "all zeros invalid", input: "00000000000000000000000000000000", want: false},
		{name: "too short", input: "4bf92f3577b34da6", want: false},
		{name: "too long", input: "4bf92f3577b34da6a3ce929d0e0e47360", want: false},
		{name: "empty", input: "", want: false},
		{name: "uppercase invalid", input: "4BF92F3577B34DA6A3CE929D0E0E4736", want: false},
		{name: "non-hex chars", input: "4bf92f3577b34da6a3ce929d0e0eXXXX", want: false},
		{name: "spaces", input: "4bf92f35 77b34da6 a3ce929d 0e0e4736", want: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := protocol.ValidTraceID(tt.input)
			if got != tt.want {
				t.Errorf("ValidTraceID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
