package goutils

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
