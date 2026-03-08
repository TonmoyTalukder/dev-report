package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/types"
)

func TestBuildPromptIncludesTaskModeInstructions(t *testing.T) {
	groups := []*types.TaskGroup{
		{
			Module:    "Npm",
			TimeSpent: 30 * time.Minute,
			Commits: []*types.Commit{
				{Message: "fix: improve installer download"},
			},
		},
	}

	balanced := BuildPrompt(groups, constants.TaskGranularityBalanced)
	detailed := BuildPrompt(groups, constants.TaskGranularityDetailed)
	granular := BuildPrompt(groups, constants.TaskGranularityGranular)

	if !strings.Contains(balanced, "avoid noisy micro-tasks") {
		t.Fatalf("expected balanced prompt guidance, got %q", balanced)
	}
	if !strings.Contains(detailed, "Prefer slightly more task separation") {
		t.Fatalf("expected detailed prompt guidance, got %q", detailed)
	}
	if !strings.Contains(granular, "Keep more distinct task rows") {
		t.Fatalf("expected granular prompt guidance, got %q", granular)
	}
}
