package ai

import "testing"

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
