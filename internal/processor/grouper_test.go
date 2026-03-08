package processor

import (
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/types"
)

func testCommit(message, path string, ts time.Time) *types.Commit {
	return &types.Commit{
		Message:   message,
		Timestamp: ts,
		Files: []types.FileChange{
			{Path: path, Added: 10},
		},
	}
}

func TestGroupCommitsDetailedKeepsMoreSeparateTasks(t *testing.T) {
	base := time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC)
	commits := []*types.Commit{
		testCommit("refresh installer flow", "npm/install.js", base),
		testCommit("improve installer retry", "npm/package.json", base.Add(30*time.Minute)),
	}

	balanced := GroupCommits(commits, constants.TaskGranularityBalanced)
	detailed := GroupCommits(commits, constants.TaskGranularityDetailed)

	if len(balanced) != 1 {
		t.Fatalf("expected balanced mode to merge nearby related module work into 1 task, got %d", len(balanced))
	}
	if len(detailed) != 2 {
		t.Fatalf("expected detailed mode to keep more separate tasks, got %d", len(detailed))
	}
}

func TestGroupCommitsDetailedStillMergesStronglyRelatedWork(t *testing.T) {
	base := time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC)
	commits := []*types.Commit{
		testCommit("refresh installer flow", "npm/install.js", base),
		testCommit("retry archive download", "npm/install.js", base.Add(45*time.Minute)),
	}

	detailed := GroupCommits(commits, constants.TaskGranularityDetailed)
	if len(detailed) != 1 {
		t.Fatalf("expected detailed mode to still merge strongly related work, got %d tasks", len(detailed))
	}
}

func TestGroupCommitsGranularKeepsEachCommitAsOwnTask(t *testing.T) {
	base := time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC)
	commits := []*types.Commit{
		testCommit("refresh installer flow", "npm/install.js", base),
		testCommit("improve installer retry", "npm/install.js", base.Add(15*time.Minute)),
	}

	granular := GroupCommits(commits, constants.TaskGranularityGranular)
	if len(granular) != len(commits) {
		t.Fatalf("expected granular mode to keep %d tasks, got %d", len(commits), len(granular))
	}
	for i, group := range granular {
		if len(group.Commits) != 1 {
			t.Fatalf("expected granular group %d to contain exactly 1 commit, got %d", i, len(group.Commits))
		}
	}
}
