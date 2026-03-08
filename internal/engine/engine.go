package engine

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dev-report/dev-report/internal/ai"
	"github.com/dev-report/dev-report/internal/config"
	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/git"
	"github.com/dev-report/dev-report/internal/processor"
	"github.com/dev-report/dev-report/internal/types"
)

// Run orchestrates the full report generation pipeline.
//
//  1. Fetch git commits
//  2. Group related commits
//  3. Calculate time weights + distribute budget
//  4. Call AI to generate task descriptions
//  5. Return a ReportOutput ready for formatting
func Run(ctx context.Context, input *types.ReportInput, cfg *config.Config) (*types.ReportOutput, error) {
	// ── Step 1: Build git fetch options ─────────────────────────────────────
	opts, err := git.BuildOptions(input)
	if err != nil {
		return nil, fmt.Errorf("build git options: %w", err)
	}

	fmt.Fprintf(os.Stderr, "  Fetching commits (user=%q, workdir=%q)…\n", input.User, input.WorkDir)
	commits, err := git.Fetch(opts)
	if err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found for the given filters — check --user, --date, or --last flags")
	}
	fmt.Fprintf(os.Stderr, "  Found %d commit(s).\n", len(commits))

	// ── Step 2: Group related commits ────────────────────────────────────────
	groups := processor.GroupCommits(commits, input.TaskMode)
	fmt.Fprintf(os.Stderr, "  Grouped into %d task(s).\n", len(groups))

	// ── Step 3: Calculate time budget ────────────────────────────────────────
	budget, err := processor.CalculateBudget(input.CheckIn, input.CheckOut, input.Adjust)
	if err != nil {
		return nil, fmt.Errorf("time budget: %w", err)
	}

	processor.AssignWeights(groups)

	if budget > 0 {
		processor.DistributeTime(groups, budget)
		fmt.Fprintf(os.Stderr, "  Time budget: %s (adjusted: %s)\n",
			processor.FormatDuration(budget), input.Adjust)
	} else {
		estimatedTotal := estimateTimeWithoutBudget(commits, groups, input.TaskMode)
		fmt.Fprintf(os.Stderr, "  No check-in/check-out provided — estimating %s total across tasks.\n",
			processor.FormatDuration(estimatedTotal))
	}

	// ── Step 4: AI task generation ───────────────────────────────────────────
	providerName := input.AIProvider
	if providerName == "" {
		providerName = cfg.AIProvider
	}
	fmt.Fprintf(os.Stderr, "  Calling AI provider: %s…\n", providerName)

	provider, err := ai.New(cfg, providerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ⚠  AI setup failed: %v\n  Falling back to commit-based report.\n", err)
		tasks := buildFallbackTasks(groups)
		return buildOutput(input, tasks, len(commits), len(groups), budget), nil
	}

	tasks, err := ai.Generate(ctx, provider, groups, input.TaskMode)
	if err != nil {
		return nil, fmt.Errorf("AI generation: %w", err)
	}

	// ── Step 5: Build output ─────────────────────────────────────────────────
	return buildOutput(input, tasks, len(commits), len(groups), budget), nil
}

// buildOutput assembles the final ReportOutput from inputs and generated tasks.
func buildOutput(input *types.ReportInput, tasks []*types.Task, commitCount, taskCount int, budget time.Duration) *types.ReportOutput {
	date := input.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	totalTime := ""
	if budget > 0 {
		totalTime = processor.FormatDuration(budget)
	} else if len(tasks) > 0 {
		var total time.Duration
		for _, t := range tasks {
			total += parseDurationStr(t.TimeSpent)
		}
		totalTime = processor.FormatDuration(total)
	}

	return &types.ReportOutput{
		Input:       input,
		Tasks:       tasks,
		CommitCount: commitCount,
		TaskCount:   taskCount,
		TotalTime:   totalTime,
		Date:        date,
		Developer:   input.User,
		CheckIn:     input.CheckIn,
		CheckOut:    input.CheckOut,
		Adjusted:    input.Adjust,
	}
}

func estimateTimeWithoutBudget(commits []*types.Commit, groups []*types.TaskGroup, taskMode string) time.Duration {
	if len(groups) == 0 {
		return 0
	}

	baselineTaskCount := len(groups)
	if taskMode != constants.TaskGranularityBalanced {
		baselineGroups := processor.GroupCommits(commits, constants.TaskGranularityBalanced)
		if len(baselineGroups) > 0 {
			baselineTaskCount = len(baselineGroups)
		}
	}

	budget := constants.DefaultTaskEstimate * time.Duration(baselineTaskCount)
	processor.DistributeTime(groups, budget)
	return budget
}

// buildFallbackTasks creates tasks without AI from commit groups.
func buildFallbackTasks(groups []*types.TaskGroup) []*types.Task {
	tasks := make([]*types.Task, len(groups))
	for i, g := range groups {
		title := "Work done"
		if len(g.Commits) > 0 {
			title = cleanMsg(g.Commits[0].Message)
		}
		tasks[i] = &types.Task{
			Number:      i + 1,
			Title:       title,
			Module:      g.Module,
			Description: fmt.Sprintf("%d commit(s) in %s", len(g.Commits), g.Module),
			TimeSpent:   processor.FormatDuration(g.TimeSpent),
			Status:      "Completed",
		}
	}
	return tasks
}

// cleanMsg strips common git prefixes from a commit message.
func cleanMsg(msg string) string {
	trimmed := strings.TrimSpace(msg)
	lower := strings.ToLower(trimmed)
	for _, kind := range constants.GitCommitTypes {
		if strings.HasPrefix(lower, kind+":") {
			return strings.TrimSpace(trimmed[len(kind)+1:])
		}
		if strings.HasPrefix(lower, kind+"(") {
			if idx := strings.Index(trimmed, "):"); idx >= 0 {
				return strings.TrimSpace(trimmed[idx+2:])
			}
		}
	}
	return trimmed
}

// parseDurationStr parses formatted strings like "2h 30m", "45m", "1h" back to Duration.
func parseDurationStr(s string) time.Duration {
	re := regexp.MustCompile(`^(?:(\d+)h)?(?:\s*(\d+)m)?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return 0
	}

	var h, m int
	if matches[1] != "" {
		h, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		m, _ = strconv.Atoi(matches[2])
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
}
