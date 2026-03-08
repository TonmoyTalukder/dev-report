package processor

import (
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
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

func TestEstimateTimeWithoutBudgetDistributesByWeight(t *testing.T) {
	groups := []*types.TaskGroup{
		{Weight: 1},
		{Weight: 2},
		{Weight: 5},
	}

	total := EstimateTimeWithoutBudget(groups)
	wantTotal := constants.DefaultTaskEstimate * time.Duration(len(groups))

	if total != wantTotal {
		t.Fatalf("expected estimated total %v, got %v", wantTotal, total)
	}

	if groups[0].TimeSpent >= groups[1].TimeSpent {
		t.Fatalf("expected group 2 to receive more time than group 1, got %v and %v", groups[1].TimeSpent, groups[0].TimeSpent)
	}
	if groups[1].TimeSpent >= groups[2].TimeSpent {
		t.Fatalf("expected group 3 to receive more time than group 2, got %v and %v", groups[2].TimeSpent, groups[1].TimeSpent)
	}

	var allocated time.Duration
	for _, group := range groups {
		allocated += group.TimeSpent
	}
	if allocated != wantTotal {
		t.Fatalf("expected allocated total %v, got %v", wantTotal, allocated)
	}

	if groups[0].TimeSpent == groups[1].TimeSpent && groups[1].TimeSpent == groups[2].TimeSpent {
		t.Fatalf("expected weighted estimation to avoid identical time allocations, got %v, %v, %v", groups[0].TimeSpent, groups[1].TimeSpent, groups[2].TimeSpent)
	}
}

func TestEstimateTimeWithoutBudgetFallsBackToDefaultWhenWeightsAreZero(t *testing.T) {
	groups := []*types.TaskGroup{{}, {}}

	total := EstimateTimeWithoutBudget(groups)
	wantPerGroup := constants.DefaultTaskEstimate
	wantTotal := wantPerGroup * time.Duration(len(groups))

	if total != wantTotal {
		t.Fatalf("expected total %v, got %v", wantTotal, total)
	}

	for i, group := range groups {
		if group.TimeSpent != wantPerGroup {
			t.Fatalf("expected group %d to receive default estimate %v, got %v", i, wantPerGroup, group.TimeSpent)
		}
	}
}

func TestDistributeTimeAllocatesMinimumFiveMinutesPerTaskWhenBudgetAllows(t *testing.T) {
	groups := []*types.TaskGroup{
		{Weight: 1},
		{Weight: 1},
		{Weight: 10},
	}

	DistributeTime(groups, 30*time.Minute)

	var total time.Duration
	for i, group := range groups {
		total += group.TimeSpent
		if group.TimeSpent < 5*time.Minute {
			t.Fatalf("expected group %d to receive at least 5m, got %v", i, group.TimeSpent)
		}
	}

	if total != 30*time.Minute {
		t.Fatalf("expected total allocation to equal budget, got %v", total)
	}
}
