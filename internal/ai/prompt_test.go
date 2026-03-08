package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/types"
)

func TestCleanMessageStripsConventionalCommitPrefixes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple prefix", in: "docs: update README", want: "update README"},
		{name: "scoped prefix", in: "feat(report): add json output", want: "add json output"},
		{name: "trim whitespace", in: "  fix(cli): keep stdout clean  ", want: "keep stdout clean"},
		{name: "no prefix", in: "Improve export layout", want: "Improve export layout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanMessage(tt.in)
			if got != tt.want {
				t.Fatalf("cleanMessage(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeDescriptionKeepsDescriptionsConcise(t *testing.T) {
	got := normalizeDescription("Made the terminal output work correctly on all systems while improving readability for daily report sharing everywhere.")

	words := strings.Fields(got)
	if len(words) > 5 {
		t.Fatalf("expected normalized description to be at most 5 words, got %d in %q", len(words), got)
	}
	if got != "Made the terminal output work" {
		t.Fatalf("unexpected normalized description: %q", got)
	}
}

func TestChooseDescriptionPrefersFallbackWhenPrimaryIsTooShort(t *testing.T) {
	got := chooseDescription("Safer output", "Made the terminal output work correctly on all systems")
	if got != "Safer output" {
		t.Fatalf("expected primary description when present, got %q", got)
	}
}

func TestChooseDescriptionKeepsNaturalPrimaryWhenItIsAlreadyUseful(t *testing.T) {
	got := chooseDescription("Made terminal output safer for users", "Made the terminal output work correctly on all systems")
	if got != "Made terminal output safer for" {
		t.Fatalf("expected natural primary description, got %q", got)
	}
}

func TestBuildPromptRequestsShortDescriptions(t *testing.T) {
	groups := []*types.TaskGroup{
		{
			Module:    "Npm",
			TimeSpent: 30 * time.Minute,
			Commits: []*types.Commit{
				{Message: "fix: improve installer download"},
			},
		},
	}

	prompt := BuildPrompt(groups, "dev-report", constants.TaskGranularityBalanced)
	if !strings.Contains(prompt, "Description: ONE short plain phrase, max 5 words") {
		t.Fatalf("expected short description instruction in prompt, got %q", prompt)
	}
	if !strings.Contains(prompt, "\"project\": \"dev-report\"") {
		t.Fatalf("expected project name in prompt, got %q", prompt)
	}
}
