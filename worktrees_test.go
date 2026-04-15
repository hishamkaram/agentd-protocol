package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestWorktrees_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		val  any
	}{
		{
			"Worktree_primary",
			Worktree{
				Path:      "/home/dev/repo",
				Branch:    "main",
				HeadSHA:   "abc1234",
				IsPrimary: true,
				Upstream:  "origin/main",
				Ahead:     2,
				Behind:    1,
			},
		},
		{
			"Worktree_secondary_locked",
			Worktree{
				Path:         "/home/dev/worktrees/repo-feat",
				Branch:       "feat/foo",
				HeadSHA:      "def5678",
				Locked:       true,
				LockedReason: "long-running experiment",
			},
		},
		{
			"Worktree_prunable",
			Worktree{
				Path:           "/home/dev/worktrees/repo-old",
				Prunable:       true,
				PrunableReason: "gitdir file points to non-existent location",
			},
		},
		{
			"GitWorktreeListRequest",
			GitWorktreeListRequest{SessionID: "s1", RequestID: "req-1"},
		},
		{
			"GitWorktreeListResponse_OK",
			GitWorktreeListResponse{
				RequestID: "req-1",
				OK:        true,
				Worktrees: []Worktree{
					{Path: "/r", Branch: "main", HeadSHA: "aaa", IsPrimary: true},
					{Path: "/r-feat", Branch: "feat", HeadSHA: "bbb"},
				},
			},
		},
		{
			"GitWorktreeListResponse_Err",
			GitWorktreeListResponse{
				RequestID: "req-1",
				OK:        false,
				Error:     "not a git repo",
				ErrorCode: GitSyncErrNotAGitRepo,
			},
		},
		{
			"GitWorktreeAddRequest_existing_branch",
			GitWorktreeAddRequest{
				SessionID: "s1",
				RequestID: "req-2",
				Path:      "/home/dev/worktrees/repo-main",
				BaseRef:   "main",
			},
		},
		{
			"GitWorktreeAddRequest_new_branch_locked",
			GitWorktreeAddRequest{
				SessionID:       "s1",
				RequestID:       "req-3",
				Path:            "/home/dev/worktrees/repo-feat",
				BaseRef:         "main",
				NewBranchName:   "feat/new",
				LockImmediately: true,
			},
		},
		{
			"GitWorktreeAddResponse_OK",
			GitWorktreeAddResponse{
				RequestID: "req-2",
				OK:        true,
				Worktree: &Worktree{
					Path:    "/home/dev/worktrees/repo-main",
					Branch:  "main",
					HeadSHA: "cafef00d",
				},
			},
		},
		{
			"GitWorktreeAddResponse_PathExists",
			GitWorktreeAddResponse{
				RequestID: "req-2",
				OK:        false,
				ErrorCode: GitWorktreeErrPathExists,
				Stderr:    "fatal: 'foo' already exists",
			},
		},
		{
			"GitWorktreeAddedPayload",
			GitWorktreeAddedPayload{
				SessionID: "s1",
				Worktree:  Worktree{Path: "/r-feat", Branch: "feat"},
			},
		},
		{
			"GitWorktreeRemoveRequest",
			GitWorktreeRemoveRequest{
				SessionID: "s1",
				RequestID: "req-4",
				Path:      "/home/dev/worktrees/repo-feat",
			},
		},
		{
			"GitWorktreeRemoveRequest_force",
			GitWorktreeRemoveRequest{
				SessionID: "s1",
				RequestID: "req-5",
				Path:      "/home/dev/worktrees/repo-feat",
				Force:     true,
			},
		},
		{
			"GitWorktreeRemoveResponse_OK",
			GitWorktreeRemoveResponse{RequestID: "req-4", OK: true},
		},
		{
			"GitWorktreeRemoveResponse_InUse",
			GitWorktreeRemoveResponse{
				RequestID: "req-4",
				OK:        false,
				ErrorCode: GitWorktreeErrInUse,
				Users:     []string{"session-abc", "session-def"},
			},
		},
		{
			"GitWorktreeRemovedPayload",
			GitWorktreeRemovedPayload{SessionID: "s1", Path: "/r-feat"},
		},
		{
			"GitWorktreeLockRequest",
			GitWorktreeLockRequest{
				SessionID: "s1",
				RequestID: "req-6",
				Path:      "/r-feat",
				Reason:    "long experiment",
			},
		},
		{
			"GitWorktreeLockResponse_OK",
			GitWorktreeLockResponse{
				RequestID: "req-6",
				OK:        true,
				Worktree: &Worktree{
					Path:         "/r-feat",
					Branch:       "feat",
					Locked:       true,
					LockedReason: "long experiment",
				},
			},
		},
		{
			"GitWorktreeUnlockRequest",
			GitWorktreeUnlockRequest{
				SessionID: "s1",
				RequestID: "req-7",
				Path:      "/r-feat",
			},
		},
		{
			"GitWorktreeUnlockResponse_OK",
			GitWorktreeUnlockResponse{
				RequestID: "req-7",
				OK:        true,
				Worktree: &Worktree{
					Path:   "/r-feat",
					Branch: "feat",
				},
			},
		},
		{
			"GitWorktreeUpdatedPayload",
			GitWorktreeUpdatedPayload{
				SessionID: "s1",
				Worktree: Worktree{
					Path:         "/r-feat",
					Branch:       "feat",
					Locked:       true,
					LockedReason: "long experiment",
				},
			},
		},
		{
			"GitWorktreePruneRequest_dry_run",
			GitWorktreePruneRequest{SessionID: "s1", RequestID: "req-8", DryRun: true},
		},
		{
			"GitWorktreePruneRequest_apply",
			GitWorktreePruneRequest{SessionID: "s1", RequestID: "req-9"},
		},
		{
			"GitWorktreePruneResponse_OK",
			GitWorktreePruneResponse{
				RequestID:    "req-8",
				OK:           true,
				DryRun:       true,
				RemovedPaths: []string{".git/worktrees/stale1", ".git/worktrees/stale2"},
				RemovedCount: 2,
			},
		},
		{
			"GitWorktreeProgressPayload",
			GitWorktreeProgressPayload{
				RequestID: "req-2",
				Op:        "add",
				Stage:     "checking_out",
				Percent:   75,
				Line:      "Checking out files: 75% (300/400)",
			},
		},
		{
			"GitWorktreeProgressPayload_indeterminate",
			GitWorktreeProgressPayload{
				RequestID: "req-2",
				Op:        "add",
				Stage:     "enumerating",
			},
		},
		{
			"SessionWorktreeSwitchRequest",
			SessionWorktreeSwitchRequest{
				SessionID: "s1",
				RequestID: "req-10",
				Path:      "/r-feat",
			},
		},
		{
			"SessionWorktreeSwitchResponse_OK",
			SessionWorktreeSwitchResponse{
				RequestID:   "req-10",
				OK:          true,
				NewWorkDir:  "/r-feat",
				NewBranch:   "feat",
				RestartedAt: 1712345678000,
			},
		},
		{
			"SessionWorktreeSwitchResponse_same",
			SessionWorktreeSwitchResponse{
				RequestID: "req-10",
				OK:        false,
				Error:     "already attached to this worktree",
				ErrorCode: "worktree_same",
			},
		},
		{
			"SessionWorktreeChangedPayload",
			SessionWorktreeChangedPayload{
				SessionID:   "s1",
				OldWorkDir:  "/r",
				NewWorkDir:  "/r-feat",
				NewBranch:   "feat",
				RestartedAt: 1712345678000,
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

func TestWorktrees_MsgConstants_NonEmpty(t *testing.T) {
	t.Parallel()
	consts := map[string]string{
		"MsgGitWorktreeList":               MsgGitWorktreeList,
		"MsgGitWorktreeListResponse":       MsgGitWorktreeListResponse,
		"MsgGitWorktreeAdd":                MsgGitWorktreeAdd,
		"MsgGitWorktreeAddResponse":        MsgGitWorktreeAddResponse,
		"MsgGitWorktreeAdded":              MsgGitWorktreeAdded,
		"MsgGitWorktreeRemove":             MsgGitWorktreeRemove,
		"MsgGitWorktreeRemoveResponse":     MsgGitWorktreeRemoveResponse,
		"MsgGitWorktreeRemoved":            MsgGitWorktreeRemoved,
		"MsgGitWorktreeLock":               MsgGitWorktreeLock,
		"MsgGitWorktreeLockResponse":       MsgGitWorktreeLockResponse,
		"MsgGitWorktreeUnlock":             MsgGitWorktreeUnlock,
		"MsgGitWorktreeUnlockResponse":     MsgGitWorktreeUnlockResponse,
		"MsgGitWorktreeUpdated":            MsgGitWorktreeUpdated,
		"MsgGitWorktreePrune":              MsgGitWorktreePrune,
		"MsgGitWorktreePruneResponse":      MsgGitWorktreePruneResponse,
		"MsgGitWorktreeProgress":           MsgGitWorktreeProgress,
		"MsgSessionWorktreeSwitch":         MsgSessionWorktreeSwitch,
		"MsgSessionWorktreeSwitchResponse": MsgSessionWorktreeSwitchResponse,
		"MsgSessionWorktreeChanged":        MsgSessionWorktreeChanged,
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
	if len(consts) != 19 {
		t.Errorf("expected 19 Msg* constants for feature 173, got %d", len(consts))
	}
}

func TestWorktrees_ErrorCodes_NonEmpty(t *testing.T) {
	t.Parallel()
	codes := map[string]string{
		"GitWorktreeErrPathExists":       GitWorktreeErrPathExists,
		"GitWorktreeErrPathInvalid":      GitWorktreeErrPathInvalid,
		"GitWorktreeErrBranchCheckedOut": GitWorktreeErrBranchCheckedOut,
		"GitWorktreeErrInUse":            GitWorktreeErrInUse,
		"GitWorktreeErrPrimary":          GitWorktreeErrPrimary,
		"GitWorktreeErrDirty":            GitWorktreeErrDirty,
		"GitWorktreeErrMissing":          GitWorktreeErrMissing,
		"GitWorktreeErrBranchInvalid":    GitWorktreeErrBranchInvalid,
		"GitWorktreesDisabled":           GitWorktreesDisabled,
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
	if len(codes) != 9 {
		t.Errorf("expected 9 error codes for feature 173, got %d", len(codes))
	}
}
