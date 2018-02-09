package goutils

import (
	"regexp"

	"github.com/hoveychen/go-utils/gomap"
)

var (
	cachedRegexp = gomap.New()
)

type Regexp struct {
	*regexp.Regexp
	err error
}

// CompileRegexp is the same as regexp.Compile(), except it cached all the
// compiled patterns for performance.
func CompileRegexp(pattern string) (*Regexp, error) {
	re := cachedRegexp.GetOrCreate(pattern, func() interface{} {
		re, err := regexp.Compile(pattern)
		return &Regexp{
			Regexp: re,
			err:    err,
		}
	}).(*Regexp)

	if re.err != nil {
		return nil, re.err
	}
	return re, nil
}

// MatchString is the same as regexp.MatchString(),
// except it use the cached version of compiled pattern.
func MatchString(pattern, s string) (matched bool, err error) {
	re, err := CompileRegexp(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

func (r *Regexp) FindNamedStringSubmatch(s string) map[string]string {
	match := r.FindStringSubmatch(s)
	if match == nil {
		return nil
	}
	ret := map[string]string{}
	for i, name := range r.SubexpNames() {
		if name != "" {
			ret[name] = match[i]
		}
	}
	return ret
}
