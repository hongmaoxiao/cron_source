package cron

import (
	"fmt"
	"testing"
)

func TestRange(t *testing.T) {
	ranges := []struct {
		expr     string
		min, max uint
		expected uint64
	}{
		// {"5", 0, 7, 1 << 5},
		// {"0", 0, 7, 1 << 0},
		// {"7", 0, 7, 1 << 7},

		{"5-5", 0, 7, 1 << 5},
		{"5-6", 0, 7, 1<<5 | 1<<6},
		{"5-7", 0, 7, 1<<5 | 1<<6 | 1<<7},

		// {"5-6/2", 0, 7, 1 << 5},
		// {"5-7/2", 0, 7, 1<<5 | 1<<7},
		// {"5-7/1", 0, 7, 1<<5 | 1<<6 | 1<<7},

		// {"*", 1, 3, 1<<1 | 1<<2 | 1<<3},
		// {"*/2", 1, 3, 1<<1 | 1<<3},
	}

	for _, c := range ranges {
		actual := getRange(c.expr, Range{c.min, c.max})
		fmt.Println("actual: ", actual)
		fmt.Println("expected: ", c.expected)
		if actual != c.expected {
			t.Errorf("%s => (expected) %d != %d (actual)", c.expr, c.expected, actual)
		}
	}
}
