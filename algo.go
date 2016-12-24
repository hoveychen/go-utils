package goutils

import (
	"sort"
)

func DedupStrings(slice []string) []string {
	m := map[string]bool{}
	ret := []string{}
	for _, s := range slice {
		if m[s] {
			continue
		} else {
			m[s] = true
			ret = append(ret, s)
		}
	}

	return ret
}

// StringSliceContains determines whether a string is contained in another slice.
// NOTE: This is just a convinient helper. It's computing time complexity is O(N), which may be a performance trap.
func StringSliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

type StringIntPair struct {
	Key   string
	Value int
}

type OrderByValue []StringIntPair

func (slice OrderByValue) Less(i, j int) bool {
	return slice[i].Value < slice[j].Value
}

func (slice OrderByValue) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice OrderByValue) Len() int {
	return len(slice)
}

func IterStringIntMap(in map[string]int) []StringIntPair {
	keys := make([]string, len(in))
	i := 0
	for key, _ := range in {
		keys[i] = key
		i++
	}

	sort.Strings(keys)
	ret := make([]StringIntPair, len(keys))
	i = 0
	for _, key := range keys {
		ret[i] = StringIntPair{key, in[key]}
		i++
	}
	return ret
}

func GenMask(slice []string) map[string]bool {
	ret := map[string]bool{}
	for _, i := range slice {
		ret[i] = true
	}
	return ret
}
