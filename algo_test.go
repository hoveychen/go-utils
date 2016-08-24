package goutils

import (
	"sort"
	"testing"
)

func TestDedupStrings(t *testing.T) {
	cases := [][]string{
		{"abc", "abc", "test", "qqq", "ppp", "abc"},
		{"abc", "ppp", "qqq", "test"},
		{"", "p", "", "qqq", "p", "abc"},
		{"", "abc", "p", "qqq"},
	}

	for i := 0; i < len(cases); i += 2 {
		ret := DedupStrings(cases[i])
		sort.Strings(ret)
		if len(ret) != len(cases[i+1]) {
			t.Error("Wrong array length. Expect", len(cases[i+1]), "Actual", len(ret))
			continue
		}

		for j := 0; j < len(ret); j++ {
			if ret[j] != cases[i+1][j] {
				t.Error("Unexpected results Expect", cases[i+1], "Actual", ret)
				break
			}
		}
	}
}
