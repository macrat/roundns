package main

import (
	"testing"
)

func TestRoundRobinStrategy(t *testing.T) {
	rr := &RoundRobinStrategy{}

	tests := []struct {
		Alives []bool
		Expect int
	}{
		{[]bool{true, true, true}, 1},
		{[]bool{true, true, true}, 2},
		{[]bool{true, true, true}, 0},
		{[]bool{true, true, true}, 1},
		{[]bool{true, true, false}, 0},
		{[]bool{true, false, false}, 0},
		{[]bool{true, false, true}, 2},
	}

	for i, tt := range tests {
		if x, err := rr.Select(tt.Alives); err != nil {
			t.Errorf("test%d: failed to select: %s", i, err)
		} else if x != tt.Expect {
			t.Errorf("test%d: unexpected selection: expected %d but got %d", i, tt.Expect, x)
		}
	}
}
