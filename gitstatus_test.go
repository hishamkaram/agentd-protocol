package protocol_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	protocol "github.com/hishamkaram/agentd-protocol"
)

func TestGitFileStatusRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   protocol.GitFileStatus
	}{
		{
			name: "modified tracked",
			in: protocol.GitFileStatus{
				Path:        "internal/session/manager.go",
				XY:          ".M",
				IsStaged:    false,
				IsUntracked: false,
				IsConflict:  false,
				IsRename:    false,
				Insertions:  7,
				Deletions:   3,
				IsBinary:    false,
				IsRedacted:  false,
				SizeBytes:   12345,
			},
		},
		{
			name: "renamed with orig path",
			in: protocol.GitFileStatus{
				Path:     "pkg/newname.go",
				OrigPath: "pkg/oldname.go",
				XY:       "R.",
				IsStaged: true,
				IsRename: true,
			},
		},
		{
			name: "redacted sensitive path",
			in: protocol.GitFileStatus{
				Path:       ".env",
				XY:         "??",
				IsUntracked: true,
				IsRedacted: true,
				SizeBytes:  256,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var out protocol.GitFileStatus
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(tt.in, out) {
				t.Errorf("roundtrip mismatch:\n in:  %+v\n out: %+v", tt.in, out)
			}
		})
	}
}

func TestGitFileStatusJSONTags(t *testing.T) {
	t.Parallel()
	f := protocol.GitFileStatus{
		Path:     "a.go",
		OrigPath: "",
		XY:       ".M",
	}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// orig_path must be omitted when empty
	if strings.Contains(s, "orig_path") {
		t.Errorf("orig_path should be omitempty, got: %s", s)
	}
	// Required fields must be present
	for _, want := range []string{
		`"path":"a.go"`,
		`"xy":".M"`,
		`"is_staged":false`,
		`"is_untracked":false`,
		`"is_conflict":false`,
		`"is_rename":false`,
		`"insertions":0`,
		`"deletions":0`,
		`"is_binary":false`,
		`"is_redacted":false`,
		`"size_bytes":0`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("JSON missing %q; got: %s", want, s)
		}
	}
}

func TestGitStatusPayloadRoundtrip(t *testing.T) {
	t.Parallel()

	in := protocol.GitStatusPayload{
		RepoRoot: "/tmp/repo",
		Files: []protocol.GitFileStatus{
			{Path: "a.go", XY: ".M", Insertions: 1, Deletions: 0},
			{Path: "b.go", XY: "??", IsUntracked: true, Insertions: 10},
		},
		TotalInsertions: 11,
		TotalDeletions:  0,
		GeneratedAt:     1712345678000,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out protocol.GitStatusPayload
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("mismatch: %+v vs %+v", in, out)
	}
	if !strings.Contains(string(data), `"generated_at":1712345678000`) {
		t.Errorf("generated_at missing from JSON: %s", data)
	}
}

func TestGitStatusPayload_Feature172_ExtendedFields(t *testing.T) {
	t.Parallel()

	// Populated path: all 5 feature-172 fields set.
	in := protocol.GitStatusPayload{
		RepoRoot:        "/tmp/repo",
		Files:           []protocol.GitFileStatus{},
		TotalInsertions: 0,
		TotalDeletions:  0,
		GeneratedAt:     1712345678000,
		Branch:          "main",
		Upstream:        "origin/main",
		Ahead:           2,
		Behind:          1,
		LastFetchedAt:   1712345678,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{
		`"branch":"main"`,
		`"upstream":"origin/main"`,
		`"ahead":2`,
		`"behind":1`,
		`"last_fetched_at":1712345678`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("populated: JSON missing %q; got: %s", want, s)
		}
	}
	var out protocol.GitStatusPayload
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("populated: mismatch\n in:  %+v\n out: %+v", in, out)
	}
}

func TestGitStatusPayload_Feature172_OmitEmpty(t *testing.T) {
	t.Parallel()

	// Older-daemon shape: feature-172 fields zero-valued. All 5 new fields
	// MUST be absent from the JSON thanks to omitempty — a backward-compat
	// requirement for older PWA clients and for drift detection.
	in := protocol.GitStatusPayload{
		RepoRoot:    "/tmp/repo",
		Files:       []protocol.GitFileStatus{},
		GeneratedAt: 1712345678000,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, forbidden := range []string{
		`"branch"`,
		`"upstream"`,
		`"ahead"`,
		`"behind"`,
		`"last_fetched_at"`,
	} {
		if strings.Contains(s, forbidden) {
			t.Errorf("omitempty broken: JSON contains %q (should be omitted when zero): %s", forbidden, s)
		}
	}
}

func TestGitStatusRequestRoundtrip(t *testing.T) {
	t.Parallel()
	in := protocol.GitStatusRequest{SessionID: "sess-1", RequestID: "req-abc"}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out protocol.GitStatusRequest
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if in != out {
		t.Errorf("mismatch: %+v vs %+v", in, out)
	}
}

func TestGitDiffRequestRoundtrip(t *testing.T) {
	t.Parallel()
	in := protocol.GitDiffRequest{SessionID: "sess-1", RequestID: "req-1", Path: "internal/x/y.go"}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out protocol.GitDiffRequest
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if in != out {
		t.Errorf("mismatch: %+v vs %+v", in, out)
	}
}

func TestGitDiffResponseRoundtrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   protocol.GitDiffResponse
	}{
		{
			name: "normal modified",
			in: protocol.GitDiffResponse{
				RequestID: "r1",
				Path:      "a.go",
				Status:    protocol.GitDiffStatusModified,
				DiffText:  "--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-old\n+new\n",
				SizeBytes: 4,
				Insertions: 1,
				Deletions:  1,
			},
		},
		{
			name: "binary",
			in: protocol.GitDiffResponse{
				RequestID: "r2",
				Path:      "image.png",
				Status:    protocol.GitDiffStatusModified,
				IsBinary:  true,
				SizeBytes: 102400,
			},
		},
		{
			name: "redacted",
			in: protocol.GitDiffResponse{
				RequestID:  "r3",
				Path:       ".env",
				Status:     protocol.GitDiffStatusModified,
				IsRedacted: true,
				SizeBytes:  128,
			},
		},
		{
			name: "truncated",
			in: protocol.GitDiffResponse{
				RequestID:   "r4",
				Path:        "big.go",
				Status:      protocol.GitDiffStatusModified,
				IsTruncated: true,
				DiffText:    "truncated body",
			},
		},
		{
			name: "conflict",
			in: protocol.GitDiffResponse{
				RequestID: "r5",
				Path:      "merge.go",
				Status:    protocol.GitDiffStatusConflict,
				DiffText:  "<<<<<<< HEAD\na\n=======\nb\n>>>>>>> feature\n",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			var out protocol.GitDiffResponse
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.in, out) {
				t.Errorf("mismatch: %+v vs %+v", tt.in, out)
			}
		})
	}
}

func TestGitNotAvailablePayloadRoundtrip(t *testing.T) {
	t.Parallel()
	tests := []protocol.GitNotAvailablePayload{
		{Reason: protocol.GitNotAvailableReasonNotARepo, Detail: "/tmp/not-a-repo has no .git"},
		{Reason: protocol.GitNotAvailableReasonGitMissing, Detail: "exec: git: not found"},
		{Reason: protocol.GitNotAvailableReasonWorkDirInvalid, Detail: "/bad/path: stat: no such file"},
	}
	for _, in := range tests {
		in := in
		t.Run(in.Reason, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(in)
			if err != nil {
				t.Fatal(err)
			}
			var out protocol.GitNotAvailablePayload
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatal(err)
			}
			if in != out {
				t.Errorf("mismatch: %+v vs %+v", in, out)
			}
		})
	}
}
