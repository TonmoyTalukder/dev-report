package processor

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/types"
)

// CalculateBudget computes the total task time budget from check-in/check-out
// and an optional adjustment string.
// checkIn and checkOut are "HH:MM". adjust is e.g. "35min", "1h40m", "2h".
// Returns the available task time (window minus adjustment).
func CalculateBudget(checkIn, checkOut, adjust string) (time.Duration, error) {
	if checkIn == "" || checkOut == "" {
		return 0, nil // no budget mode (only commit count used)
	}

	inTime, err := time.Parse("15:04", checkIn)
	if err != nil {
		return 0, fmt.Errorf("invalid check-in time %q: %w", checkIn, err)
	}
	outTime, err := time.Parse("15:04", checkOut)
	if err != nil {
		return 0, fmt.Errorf("invalid check-out time %q: %w", checkOut, err)
	}

	window := outTime.Sub(inTime)
	if window <= 0 {
		return 0, fmt.Errorf("check-out must be after check-in")
	}

	var adj time.Duration
	if adjust != "" {
		adj, err = ParseAdjust(adjust)
		if err != nil {
			return 0, fmt.Errorf("invalid adjust value %q: %w", adjust, err)
		}
	}

	budget := window - adj
	if budget <= 0 {
		return 0, fmt.Errorf("adjusted time budget is zero or negative")
	}
	return budget, nil
}

// ParseAdjust parses an adjustment string like "35min", "1h40m", "2h", "30m", "90min".
func ParseAdjust(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	// Patterns: "1h40m", "1h", "40m", "35min", "90min"
	re := regexp.MustCompile(`^(?:(\d+)h)?(?:(\d+)m(?:in)?)?$`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("unrecognised format (use e.g. 35min, 1h40m, 2h)")
	}

	var total time.Duration
	if m[1] != "" {
		h, _ := strconv.Atoi(m[1])
		total += time.Duration(h) * time.Hour
	}
	if m[2] != "" {
		min, _ := strconv.Atoi(m[2])
		total += time.Duration(min) * time.Minute
	}
	if total == 0 {
		return 0, fmt.Errorf("zero duration — check format (use e.g. 35min, 1h40m)")
	}
	return total, nil
}

// AssignWeights calculates a weight for each TaskGroup based on:
// commit count, commit time spread, file count, lines changed, file complexity,
// and commit importance prefix.
func AssignWeights(groups []*types.TaskGroup) {
	for _, g := range groups {
		g.Weight = calculateWeight(g)
	}
}

func calculateWeight(g *types.TaskGroup) float64 {
	weight := 0.0

	// Factor 1: commit count (each commit = 2 points)
	weight += float64(len(g.Commits)) * 2.0

	// Factor 2: commit time spread (1 point per 30 min gap)
	spread := g.LatestTime().Sub(g.EarliestTime()).Minutes()
	weight += spread / 30.0

	// Factor 3: unique file count (1.5 points per file)
	weight += float64(g.TotalFiles()) * 1.5

	// Factor 4: lines changed (1 point per 50 lines)
	weight += float64(g.TotalLines()) / 50.0

	// Factor 5: file complexity boost
	for _, c := range g.Commits {
		for _, f := range c.Files {
			ext := fileExt(f.Path)
			if boost, ok := constants.ComplexityBoost[ext]; ok {
				weight += boost
			}
		}
	}

	// Factor 6: commit importance multiplier (average across commits)
	impTotal := 0.0
	for _, c := range g.Commits {
		impTotal += commitImportance(c.Message)
	}
	avgImp := impTotal / float64(len(g.Commits))
	weight *= avgImp

	// Minimum weight so every task gets some time
	if weight < 1.0 {
		weight = 1.0
	}

	return weight
}

// DistributeTime assigns a proportional TimeSpent duration to each group
// so that the sum equals budget. Values are rounded to the nearest 5 minutes.
// The last group absorbs any rounding remainder to ensure exact total.
func DistributeTime(groups []*types.TaskGroup, budget time.Duration) {
	if len(groups) == 0 || budget == 0 {
		return
	}

	totalWeight := 0.0
	for _, g := range groups {
		totalWeight += g.Weight
	}
	if totalWeight <= 0 {
		return
	}

	totalBudgetMinutes := int(math.Round(budget.Minutes()))
	allocatedMinutes := 0

	for i, g := range groups {
		isLast := i == len(groups)-1
		remainingMinutes := totalBudgetMinutes - allocatedMinutes
		if remainingMinutes < 0 {
			remainingMinutes = 0
		}

		if isLast {
			// Last group gets the remainder to avoid drift from rounding
			g.TimeSpent = time.Duration(remainingMinutes) * time.Minute
		} else {
			rawMinutes := (g.Weight / totalWeight) * float64(totalBudgetMinutes)
			rounded := int(roundToNearest5(rawMinutes))
			if rounded > remainingMinutes {
				rounded = remainingMinutes
			}
			g.TimeSpent = time.Duration(rounded) * time.Minute
			allocatedMinutes += rounded
		}
	}
}

// FormatDuration converts a time.Duration into a human-readable string like
// "2h 30m", "45m", "1h".
func FormatDuration(d time.Duration) string {
	total := int(math.Round(d.Minutes()))
	if total <= 0 {
		return "< 1m"
	}
	h := total / 60
	m := total % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}

// roundToNearest5 rounds minutes to the nearest 5-minute increment.
func roundToNearest5(minutes float64) float64 {
	return math.Round(minutes/5) * 5
}

// fileExt returns the lowercase file extension including the dot.
func fileExt(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx < 0 {
		return ""
	}
	return strings.ToLower(path[idx:])
}

// commitImportance returns the importance multiplier for a commit message prefix.
func commitImportance(message string) float64 {
	lower := strings.ToLower(message)
	for prefix, mult := range constants.ImportanceMultiplier {
		if strings.HasPrefix(lower, prefix+":") || strings.HasPrefix(lower, prefix+"(") {
			return mult
		}
	}
	return 1.0 // default multiplier
}
