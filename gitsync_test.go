package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGitSync_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
	}{
		{
			"GitBranch",
			GitBranch{
				Name:          "feature/foo",
				IsCurrent:     true,
				Upstream:      "origin/feature/foo",
				Ahead:         2,
				Behind:        1,
				LastCommitAt:  1712345678,
				LastCommitSha: "abc1234",
			},
		},
		{
			"GitBranchListRequest",
			GitBranchListRequest{SessionID: "s1", RequestID: "req-1"},
		},
		{
			"GitBranchListResponse_OK",
			GitBranchListResponse{
				RequestID:     "req-1",
				OK:            true,
				CurrentBranch: "main",
				Branches: []GitBranch{
					{Name: "main", IsCurrent: true, Ahead: 0, Behind: 0, LastCommitAt: 1712345678},
					{Name: "feature", IsCurrent: false, Ahead: 2, Behind: 0, LastCommitAt: 1712345000},
				},
			},
		},
		{
			"GitBranchListResponse_Err",
			GitBranchListResponse{
				RequestID: "req-1",
				OK:        false,
				Error:     "not a git repository",
				ErrorCode: GitSyncErrNotAGitRepo,
			},
		},
		{
			"GitBranchSwitchRequest_clean",
			GitBranchSwitchRequest{SessionID: "s1", RequestID: "req-2", Branch: "feature", StashDirty: false},
		},
		{
			"GitBranchSwitchRequest_dirty",
			GitBranchSwitchRequest{SessionID: "s1", RequestID: "req-3", Branch: "feature", StashDirty: true},
		},
		{
			"GitBranchSwitchResponse_OK_clean",
			GitBranchSwitchResponse{RequestID: "req-2", OK: true, NewBranch: "feature"},
		},
		{
			"GitBranchSwitchResponse_OK_stashed",
			GitBranchSwitchResponse{RequestID: "req-3", OK: true, NewBranch: "feature", StashRef: "stash@{0}"},
		},
		{
			"GitBranchSwitchResponse_DirtyRequired",
			GitBranchSwitchResponse{
				RequestID:     "req-4",
				OK:            false,
				ErrorCode:     GitSyncErrDirtyWorkTree,
				DirtyRequired: true,
				Stderr:        "error: Your local changes would be overwritten",
			},
		},
		{
			"GitFetchRequest_default_remote",
			GitFetchRequest{SessionID: "s1", RequestID: "req-5"},
		},
		{
			"GitFetchRequest_explicit_remote_prune",
			GitFetchRequest{SessionID: "s1", RequestID: "req-6", Remote: "origin", Prune: true},
		},
		{
			"GitFetchResponse",
			GitFetchResponse{RequestID: "req-5", OK: true, FetchedAt: 1712345999},
		},
		{
			"GitPullRequest_ff_only",
			GitPullRequest{SessionID: "s1", RequestID: "req-7", Rebase: false},
		},
		{
			"GitPullRequest_rebase",
			GitPullRequest{SessionID: "s1", RequestID: "req-8", Rebase: true},
		},
		{
			"GitPullResponse_OK",
			GitPullResponse{RequestID: "req-7", OK: true, NewHead: "def5678", FilesChanged: 3},
		},
		{
			"GitPushRequest_plain",
			GitPushRequest{SessionID: "s1", RequestID: "req-9"},
		},
		{
			"GitPushRequest_set_upstream",
			GitPushRequest{SessionID: "s1", RequestID: "req-10", SetUpstream: true},
		},
		{
			"GitPushRequest_force_with_lease",
			GitPushRequest{SessionID: "s1", RequestID: "req-11", ForceWithLease: true},
		},
		{
			"GitPushRequest_raw_force",
			GitPushRequest{SessionID: "s1", RequestID: "req-12", Force: true},
		},
		{
			"GitPushResponse_OK",
			GitPushResponse{RequestID: "req-9", OK: true, PushedRef: "refs/heads/main -> origin/main"},
		},
		{
			"GitPushResponse_NonFastForward",
			GitPushResponse{
				RequestID: "req-9",
				OK:        false,
				ErrorCode: GitSyncErrNonFastForward,
				Stderr:    "! [rejected] main -> main (non-fast-forward)",
			},
		},
		{
			"GitStashEntry",
			GitStashEntry{
				Ref:       "stash@{1}",
				Index:     1,
				CommitSHA: "abc1234def5678",
				CreatedAt: 1712346000,
				Subject:   "On main: agentd-manual-stash-1712346000",
				Branch:    "main",
			},
		},
		{
			"GitStashListRequest",
			GitStashListRequest{SessionID: "s1", RequestID: "req-stash-list"},
		},
		{
			"GitStashListResponse_OK",
			GitStashListResponse{
				RequestID: "req-stash-list",
				OK:        true,
				Stashes: []GitStashEntry{
					{Ref: "stash@{0}", Index: 0, CommitSHA: "abc1234", CreatedAt: 1712346000, Subject: "On main: agentd-manual-stash-1712346000", Branch: "main"},
				},
			},
		},
		{
			"GitStashListResponse_Err",
			GitStashListResponse{RequestID: "req-stash-list", OK: false, ErrorCode: GitSyncErrNotAGitRepo, Error: "not a git repository"},
		},
		{
			"GitStashPushRequest",
			GitStashPushRequest{SessionID: "s1", RequestID: "req-stash-push"},
		},
		{
			"GitStashPushResponse_Stashed",
			GitStashPushResponse{RequestID: "req-stash-push", OK: true, Stashed: true, StashRef: "stash@{0}"},
		},
		{
			"GitStashPushResponse_Clean",
			GitStashPushResponse{RequestID: "req-stash-push", OK: true, Stashed: false},
		},
		{
			"GitStashApplyRequest",
			GitStashApplyRequest{SessionID: "s1", RequestID: "req-stash-apply", Ref: "stash@{0}", AllowDirty: true},
		},
		{
			"GitStashApplyResponse_OK",
			GitStashApplyResponse{RequestID: "req-stash-apply", OK: true, StashRef: "stash@{0}"},
		},
		{
			"GitStashPopRequest",
			GitStashPopRequest{SessionID: "s1", RequestID: "req-stash-pop", Ref: "0"},
		},
		{
			"GitStashPopResponse_OK",
			GitStashPopResponse{RequestID: "req-stash-pop", OK: true, StashRef: "stash@{0}", Removed: true},
		},
		{
			"GitStashPopResponse_Conflict",
			GitStashPopResponse{RequestID: "req-stash-pop", OK: false, ErrorCode: GitSyncErrMergeConflict, StashRef: "stash@{0}", Removed: false},
		},
		{
			"GitStashDropRequest",
			GitStashDropRequest{SessionID: "s1", RequestID: "req-stash-drop", Ref: "stash@{0}"},
		},
		{
			"GitStashDropResponse_OK",
			GitStashDropResponse{RequestID: "req-stash-drop", OK: true, StashRef: "stash@{0}", Removed: true},
		},
		{
			"GitSyncProgressPayload_fetch",
			GitSyncProgressPayload{
				RequestID: "req-5",
				Op:        "fetch",
				Stage:     "receiving",
				Percent:   42,
				Line:      "Receiving objects:  42% (1260/3000)",
			},
		},
		{
			"GitSyncProgressPayload_indeterminate",
			GitSyncProgressPayload{
				RequestID: "req-5",
				Op:        "fetch",
				Stage:     "enumerating",
			},
		},
		{
			"GitSyncCancelRequest",
			GitSyncCancelRequest{SessionID: "s1", RequestID: "req-cancel-1", TargetID: "req-5"},
		},
		{
			"GitSyncCancelResponse_OK",
			GitSyncCancelResponse{RequestID: "req-cancel-1", TargetID: "req-5", OK: true},
		},
		{
			"GitSyncCancelResponse_NotFound",
			GitSyncCancelResponse{
				RequestID: "req-cancel-1",
				TargetID:  "req-unknown",
				OK:        false,
				ErrorCode: GitSyncErrNotFound,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(tc.val)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			// Round-trip through a fresh value of the same concrete type.
			out := reflect.New(reflect.TypeOf(tc.val)).Interface()
			if err := json.Unmarshal(raw, out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := reflect.ValueOf(out).Elem().Interface()
			if !reflect.DeepEqual(tc.val, got) {
				t.Errorf("round trip mismatch:\n  want: %#v\n  got:  %#v", tc.val, got)
			}
		})
	}
}

func TestGitSync_MsgConstants_NonEmpty(t *testing.T) {
	t.Parallel()
	// Catches accidental deletion or typo in a constant literal.
	consts := map[string]string{
		"MsgGitBranchList":           MsgGitBranchList,
		"MsgGitBranchListResponse":   MsgGitBranchListResponse,
		"MsgGitBranchSwitch":         MsgGitBranchSwitch,
		"MsgGitBranchSwitchResponse": MsgGitBranchSwitchResponse,
		"MsgGitFetch":                MsgGitFetch,
		"MsgGitFetchResponse":        MsgGitFetchResponse,
		"MsgGitPull":                 MsgGitPull,
		"MsgGitPullResponse":         MsgGitPullResponse,
		"MsgGitPush":                 MsgGitPush,
		"MsgGitPushResponse":         MsgGitPushResponse,
		"MsgGitSyncProgress":         MsgGitSyncProgress,
		"MsgGitSyncCancel":           MsgGitSyncCancel,
		"MsgGitSyncCancelResponse":   MsgGitSyncCancelResponse,
		"MsgGitStashList":            MsgGitStashList,
		"MsgGitStashListResponse":    MsgGitStashListResponse,
		"MsgGitStashPush":            MsgGitStashPush,
		"MsgGitStashPushResponse":    MsgGitStashPushResponse,
		"MsgGitStashApply":           MsgGitStashApply,
		"MsgGitStashApplyResponse":   MsgGitStashApplyResponse,
		"MsgGitStashPop":             MsgGitStashPop,
		"MsgGitStashPopResponse":     MsgGitStashPopResponse,
		"MsgGitStashDrop":            MsgGitStashDrop,
		"MsgGitStashDropResponse":    MsgGitStashDropResponse,
	}
	seen := map[string]string{}
	for name, val := range consts {
		if val == "" {
			t.Errorf("%s is empty", name)
		}
		if prev, dup := seen[val]; dup {
			t.Errorf("%s and %s share literal %q", prev, name, val)
		}
		seen[val] = name
	}
}

func TestGitSync_ErrorCodes_NonEmpty(t *testing.T) {
	t.Parallel()
	codes := map[string]string{
		"GitSyncErrAuthFailed":     GitSyncErrAuthFailed,
		"GitSyncErrSSHPromptHang":  GitSyncErrSSHPromptHang,
		"GitSyncErrNonFastForward": GitSyncErrNonFastForward,
		"GitSyncErrMergeConflict":  GitSyncErrMergeConflict,
		"GitSyncErrDirtyWorkTree":  GitSyncErrDirtyWorkTree,
		"GitSyncErrNoUpstream":     GitSyncErrNoUpstream,
		"GitSyncErrNotAGitRepo":    GitSyncErrNotAGitRepo,
		"GitSyncErrNetwork":        GitSyncErrNetwork,
		"GitSyncErrCanceled":       GitSyncErrCanceled,
		"GitSyncErrTimeout":        GitSyncErrTimeout,
		"GitSyncErrLockedIndex":    GitSyncErrLockedIndex,
		"GitSyncErrNotFound":       GitSyncErrNotFound,
		"GitSyncErrInternal":       GitSyncErrInternal,
	}
	seen := map[string]string{}
	for name, val := range codes {
		if val == "" {
			t.Errorf("%s is empty", name)
		}
		if prev, dup := seen[val]; dup {
			t.Errorf("%s and %s share literal %q", prev, name, val)
		}
		seen[val] = name
	}
}
