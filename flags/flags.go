package flags

import (
	"flag"
	"strings"
	"time"
)

var (
	stringPtrs   = map[string]*string{}
	boolPtrs     = map[string]*bool{}
	durationPtrs = map[string]*time.Duration{}
	float64Ptrs  = map[string]*float64{}
	intPtrs      = map[string]*int{}
	slicePtrs    = map[string]*[]string{}
)

type sliceContainer struct {
	S *[]string
}

func (s *sliceContainer) String() string {
	return strings.Join(*s.S, ",")
}

func (s *sliceContainer) Set(value string) error {
	*s.S = strings.Split(value, ",")
	return nil
}

func String(name string, value string, usage string) *string {
	if _, exists := stringPtrs[name]; !exists {
		stringPtrs[name] = flag.String(name, value, usage)
	}

	return stringPtrs[name]
}

func Int(name string, value int, usage string) *int {
	if _, exists := intPtrs[name]; !exists {
		intPtrs[name] = flag.Int(name, value, usage)
	}

	return intPtrs[name]
}

func Bool(name string, value bool, usage string) *bool {
	if _, exists := boolPtrs[name]; !exists {
		boolPtrs[name] = flag.Bool(name, value, usage)
	}

	return boolPtrs[name]
}

func Float64(name string, value float64, usage string) *float64 {
	if _, exists := float64Ptrs[name]; !exists {
		float64Ptrs[name] = flag.Float64(name, value, usage)
	}

	return float64Ptrs[name]
}

func Slice(name string, value []string, usage string) *[]string {
	if _, exists := slicePtrs[name]; !exists {
		var s []string
		if value == nil {
			s = []string{}
		} else {
			s = value
		}
		slicePtrs[name] = &s

		container := sliceContainer{S: &s}
		flag.Var(&container, name, usage)
	}
	return slicePtrs[name]
}

func Duration(name string, value time.Duration, usage string) *time.Duration {
	if _, exists := float64Ptrs[name]; !exists {
		durationPtrs[name] = flag.Duration(name, value, usage)
	}

	return durationPtrs[name]
}
