package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/processor"
	"github.com/dev-report/dev-report/internal/types"
)

// promptGroup is the JSON structure sent to the AI per task group.
type promptGroup struct {
	GroupID   int      `json:"groupId"`
	Project   string   `json:"project"`
	Messages  []string `json:"commitMessages"`
	Files     []string `json:"changedFiles"`
	TimeSpent string   `json:"timeSpent"`
}

// AITask is the expected JSON structure returned by the AI per task.
type AITask struct {
	GroupID     int    `json:"groupId"`
	Task        string `json:"task"`
	Project     string `json:"project"`
	Module      string `json:"module"`
	Description string `json:"description"`
	TimeSpent   string `json:"timeSpent"`
	Status      string `json:"status"`
}

// BuildPrompt constructs the full prompt string to send to the AI.
func BuildPrompt(groups []*types.TaskGroup, projectName, taskMode string) string {
	var pg []promptGroup
	for i, g := range groups {
		messages := make([]string, 0, len(g.Commits))
		fileSet := map[string]bool{}
		for _, c := range g.Commits {
			msg := cleanMessage(c.Message)
			if msg != "" {
				messages = append(messages, msg)
			}
			for _, f := range c.Files {
				fileSet[f.Path] = true
			}
		}
		files := make([]string, 0, len(fileSet))
		for f := range fileSet {
			files = append(files, f)
		}

		pg = append(pg, promptGroup{
			GroupID:   i + 1,
			Project:   projectName,
			Messages:  messages,
			Files:     files,
			TimeSpent: processor.FormatDuration(g.TimeSpent),
		})
	}

	dataJSON, _ := json.MarshalIndent(pg, "", "  ")
	granularityRule := "Keep the report natural and concise. Merge obviously related work into one clear task and avoid noisy micro-tasks."
	if taskMode == constants.TaskGranularityDetailed {
		granularityRule = "Prefer slightly more task separation when groups represent clearly different outcomes, but keep the report natural and avoid repetitive or tiny micro-tasks."
	} else if taskMode == constants.TaskGranularityGranular {
		granularityRule = "Keep more distinct task rows and preserve separate work items when the commit evidence points to different outcomes, but still avoid awkward wording or trivial filler tasks."
	}

	return fmt.Sprintf(`You are writing a professional developer work report.

For each commit group below, write one report row.

RULES — read carefully:
1. Task title: plain English, simple words, max 8 words. No file names, no code.
2. Description: ONE short plain phrase, max 5 words. No technical terms, no file names, no variable names.
3. Write as if explaining to a non-technical manager who does not know programming.
4. Project: use the provided project name exactly.
5. TimeSpent: copy the provided value EXACTLY — do not change it.
6. Status: always "Completed".
7. groupId: copy the provided groupId exactly.
8. %s

EXAMPLES:
  BAD task:  "feat: add return_amount col in doctorSummary.tsx"
  GOOD task: "Added return amount to doctor summary"

  BAD description: "Refactored DI container and fixed null pointer in UserService.java"
  GOOD description: "Fixed user loading issue"

  BAD task:  "fix: useEffect hook cleanup in StoreComponent"
  GOOD task: "Fixed a display issue in the store screen"

Return ONLY a valid JSON array — no markdown, no explanation, nothing else.
Format:
[{"groupId":1,"task":"...","project":"...","description":"...","timeSpent":"...","status":"Completed"}]

Commit groups:
%s`, granularityRule, string(dataJSON))
}

// ParseResponse extracts the JSON array from the AI response, handling
// cases where the AI wraps it in markdown code fences.
func ParseResponse(raw string) ([]AITask, error) {
	raw = strings.TrimSpace(raw)

	// Strip ```json ... ``` or ``` ... ``` fences
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(\\[.*?\\])\\s*```")
	if m := re.FindStringSubmatch(raw); m != nil {
		raw = m[1]
	}

	// Find the JSON array even if there's surrounding text
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var tasks []AITask
	if err := json.Unmarshal([]byte(raw), &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w\nRaw: %s", err, truncate(raw, 500))
	}
	return tasks, nil
}

// Generate calls the AI provider, parses the response, and returns tasks.
// On failure it falls back to generating tasks directly from the commit data.
func Generate(ctx context.Context, p Provider, groups []*types.TaskGroup, projectName, taskMode string) ([]*types.Task, error) {
	prompt := BuildPrompt(groups, projectName, taskMode)

	raw, err := p.Generate(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠  AI provider (%s) failed: %v\n   Falling back to commit-based report.\n", p.Name(), err)
		return fallback(groups, projectName), nil
	}

	aiTasks, err := ParseResponse(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠  AI response parse failed: %v\n   Falling back to commit-based report.\n", err)
		return fallback(groups, projectName), nil
	}

	// Map AI tasks back by groupId, preserve order
	taskMap := map[int]*AITask{}
	for i := range aiTasks {
		taskMap[aiTasks[i].GroupID] = &aiTasks[i]
	}

	tasks := make([]*types.Task, 0, len(groups))
	for i, g := range groups {
		groupID := i + 1
		t := &types.Task{
			Number:    groupID,
			Project:   projectName,
			TimeSpent: processor.FormatDuration(g.TimeSpent),
			Status:    "Completed",
		}
		if ai, ok := taskMap[groupID]; ok {
			t.Title = ai.Task
			t.Description = chooseDescription(ai.Description, fallbackDesc(g))
			if ai.Project != "" {
				t.Project = ai.Project
			} else if ai.Module != "" {
				t.Project = ai.Module
			}
		} else {
			// AI didn't return this group — use fallback for it
			t.Title = fallbackTitle(g)
			t.Description = normalizeDescription(fallbackDesc(g))
		}
		if t.Description == "" {
			t.Description = normalizeDescription(fallbackDesc(g))
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// fallback builds tasks directly from commit data without AI.
func fallback(groups []*types.TaskGroup, projectName string) []*types.Task {
	tasks := make([]*types.Task, len(groups))
	for i, g := range groups {
		tasks[i] = &types.Task{
			Number:      i + 1,
			Title:       fallbackTitle(g),
			Project:     projectName,
			Description: normalizeDescription(fallbackDesc(g)),
			TimeSpent:   processor.FormatDuration(g.TimeSpent),
			Status:      "Completed",
		}
	}
	return tasks
}

// fallbackTitle creates a cleaned title from the first commit message.
func fallbackTitle(g *types.TaskGroup) string {
	if len(g.Commits) == 0 {
		return "Work done"
	}
	msg := cleanMessage(g.Commits[0].Message)
	if msg == "" {
		return "Work done"
	}
	words := strings.Fields(msg)
	if len(words) > 8 {
		words = words[:8]
	}
	title := strings.Join(words, " ")
	if len(title) > 0 {
		title = strings.ToUpper(string(title[0])) + title[1:]
	}
	return title
}

// fallbackDesc creates a description listing all commit messages.
func fallbackDesc(g *types.TaskGroup) string {
	msgs := make([]string, 0, len(g.Commits))
	for _, c := range g.Commits {
		m := cleanMessage(c.Message)
		if m != "" {
			msgs = append(msgs, m)
		}
	}
	if len(msgs) == 0 {
		return "Development work completed."
	}
	return strings.Join(msgs, "; ")
}

func chooseDescription(primary, backup string) string {
	if strings.TrimSpace(primary) != "" {
		return normalizeDescription(primary)
	}
	return normalizeDescription(backup)
}

func normalizeDescription(desc string) string {
	cleaned := strings.Join(strings.Fields(desc), " ")
	if cleaned == "" {
		return "Completed planned project work"
	}

	if idx := strings.IndexAny(cleaned, ".!?"); idx >= 0 {
		cleaned = strings.TrimSpace(cleaned[:idx])
	}

	words := strings.Fields(cleaned)
	if len(words) > 5 {
		words = words[:5]
	}

	cleaned = strings.Join(words, " ")
	cleaned = strings.Trim(cleaned, " .,:;!?")
	if cleaned == "" {
		return "Completed planned project work"
	}
	cleaned = strings.ToUpper(string(cleaned[0])) + cleaned[1:]
	return cleaned
}

func descriptionWordCount(desc string) int {
	return len(strings.Fields(desc))
}

// cleanMessage strips git prefix tags from a commit message.
func cleanMessage(msg string) string {
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
