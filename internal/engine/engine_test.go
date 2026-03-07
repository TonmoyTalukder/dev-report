package engine

import "testing"

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
