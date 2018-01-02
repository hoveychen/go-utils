package goutils

import "regexp"

var (
	cachedRegexp = map[string]*regexp.Regexp{}
	cachedError  = map[string]error{}
)

// CompileRegexp is the same as regexp.Compile(), except it cached all the
// compiled patterns for performance.
func CompileRegexp(pattern string) (*regexp.Regexp, error) {
	if cachedRegexp[pattern] != nil {
		return cachedRegexp[pattern], nil
	}
	if cachedError[pattern] != nil {
		return nil, cachedError[pattern]
	}
	re, err := regexp.Compile(pattern)
	cachedRegexp[pattern] = re
	cachedError[pattern] = err
	return re, err
}

// MatchStringRegexp is the same as regexp.MatchString(),
// except it use the cached version of compiled pattern.
func MatchStringRegexp(pattern, s string) (matched bool, err error) {
	re, err := CompileRegexp(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}
