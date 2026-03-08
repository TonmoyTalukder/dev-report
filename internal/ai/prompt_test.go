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
	if len(words) > 15 {
		t.Fatalf("expected normalized description to be at most 15 words, got %d in %q", len(words), got)
	}
	if got != "Made the terminal output work correctly on all systems while improving readability for daily report" {
		t.Fatalf("unexpected normalized description: %q", got)
	}
}

func TestChooseDescriptionPrefersFallbackWhenPrimaryIsTooShort(t *testing.T) {
	got := chooseDescription("Safer output", "Made the terminal output work correctly on all systems")
	if got != "Made the terminal output work correctly on all systems" {
		t.Fatalf("expected fallback description, got %q", got)
	}
}

func TestChooseDescriptionKeepsNaturalPrimaryWhenItIsAlreadyUseful(t *testing.T) {
	got := chooseDescription("Made terminal output safer for users", "Made the terminal output work correctly on all systems")
	if got != "Made terminal output safer for users" {
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

	prompt := BuildPrompt(groups, constants.TaskGranularityBalanced)
	if !strings.Contains(prompt, "Description: ONE plain sentence, ideally around 10 to 15 simple words, and never more than 15 words") {
		t.Fatalf("expected short description instruction in prompt, got %q", prompt)
	}
}
