package engine

import (
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/processor"
	"github.com/dev-report/dev-report/internal/types"
)

func TestCleanMsgStripsConventionalCommitPrefixes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple prefix", in: "fix: handle nil pointer", want: "handle nil pointer"},
		{name: "scoped prefix", in: "feat(api): add user endpoint", want: "add user endpoint"},
		{name: "uppercase scope preserved", in: "refactor(UI): simplify modal layout", want: "simplify modal layout"},
		{name: "no prefix", in: "Improve report generation", want: "Improve report generation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanMsg(tt.in)
			if got != tt.want {
				t.Fatalf("cleanMsg(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEstimateTimeWithoutBudgetUsesBalancedBaselineForGranularMode(t *testing.T) {
	base := time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC)
	commits := []*types.Commit{
		{Message: "refresh installer flow", Timestamp: base, Files: []types.FileChange{{Path: "npm/install.js", Added: 10}}},
		{Message: "improve installer retry", Timestamp: base.Add(15 * time.Minute), Files: []types.FileChange{{Path: "npm/install.js", Added: 8}}},
		{Message: "publish package metadata", Timestamp: base.Add(30 * time.Minute), Files: []types.FileChange{{Path: "npm/package.json", Added: 5}}},
	}

	balancedGroups := processor.GroupCommits(commits, constants.TaskGranularityBalanced)
	granularGroups := processor.GroupCommits(commits, constants.TaskGranularityGranular)
	processor.AssignWeights(granularGroups)

	got := estimateTimeWithoutBudget(commits, granularGroups, constants.TaskGranularityGranular)
	want := constants.DefaultTaskEstimate * time.Duration(len(balancedGroups))

	if got != want {
		t.Fatalf("expected granular no-budget estimate %v to match balanced baseline %v", got, want)
	}

	var total time.Duration
	for _, group := range granularGroups {
		total += group.TimeSpent
	}
	if total != want {
		t.Fatalf("expected granular distributed total %v, got %v", want, total)
	}
}
