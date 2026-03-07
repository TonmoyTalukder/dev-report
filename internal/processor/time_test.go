package processor

import (
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/types"
)

func TestDistributeTimeNeverAllocatesNegativeRemainder(t *testing.T) {
	groups := []*types.TaskGroup{
		{Weight: 1},
		{Weight: 1},
		{Weight: 1},
	}

	DistributeTime(groups, 5*time.Minute)

	var total time.Duration
	for i, group := range groups {
		if group.TimeSpent < 0 {
			t.Fatalf("group %d received negative time: %v", i, group.TimeSpent)
		}
		total += group.TimeSpent
	}

	if total != 5*time.Minute {
		t.Fatalf("expected total allocation to equal budget, got %v", total)
	}
}

func TestDistributeTimeCapsRoundedAllocationToRemainingBudget(t *testing.T) {
	groups := []*types.TaskGroup{
		{Weight: 1},
		{Weight: 1},
		{Weight: 1},
	}

	DistributeTime(groups, 10*time.Minute)

	if groups[0].TimeSpent != 5*time.Minute {
		t.Fatalf("expected first group to round to 5m, got %v", groups[0].TimeSpent)
	}
	if groups[1].TimeSpent != 5*time.Minute {
		t.Fatalf("expected second group to consume remaining 5m, got %v", groups[1].TimeSpent)
	}
	if groups[2].TimeSpent != 0 {
		t.Fatalf("expected final group to get 0m remainder, got %v", groups[2].TimeSpent)
	}
}
