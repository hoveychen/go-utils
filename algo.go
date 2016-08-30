package goutils

import (
	"sort"
)

func DedupStrings(slice []string) []string {
	m := map[string]bool{}
	for _, i := range slice {
		m[i] = true
	}

	ret := make([]string, len(m))
	i := 0
	for v, _ := range m {
		ret[i] = v
		i++
	}

	return ret
}

type StringIntPair struct {
	Key   string
	Value int
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
