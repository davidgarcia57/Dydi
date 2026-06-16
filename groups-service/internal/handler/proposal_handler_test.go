package handler

import "testing"

// quorumReached is the core democratic primitive: a proposal passes when
// yes-votes are at least half of the frozen electorate. These cases pin the
// boundaries (especially ties on even electorates, which must pass at exactly 50%).
func TestQuorumReached(t *testing.T) {
	cases := []struct {
		name      string
		approvals int
		members   int
		want      bool
	}{
		{"no members yet", 1, 0, false},
		{"solo group, one yes", 1, 1, true},
		{"two members, one yes is exactly 50%", 1, 2, true},
		{"two members, zero yes", 0, 2, false},
		{"three members, one yes (<50%)", 1, 3, false},
		{"three members, two yes (>50%)", 2, 3, true},
		{"four members, two yes (exactly 50%)", 2, 4, true},
		{"eight members, three yes (<50%)", 3, 8, false},
		{"eight members, four yes (exactly 50%)", 4, 8, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := quorumReached(c.approvals, c.members); got != c.want {
				t.Fatalf("quorumReached(%d, %d) = %v, want %v",
					c.approvals, c.members, got, c.want)
			}
		})
	}
}
