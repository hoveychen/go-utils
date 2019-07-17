// Package flags provide a global flags cache and extends types of flags.
package flags

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	ptrs = map[string]interface{}{}
)

// ValidateNonZero checks the flags non-zero after parsing.
func ValidateNonZero(names ...string) error {
	for _, name := range names {
		ptr, exists := ptrs[name]
		if !exists {
			return fmt.Errorf("%s is not defined", name)
		}

		check := false
		switch ptr.(type) {
		case *bool:
			// It's trivial to check a bool, since it makes the flag no sense(always true).
			check = *ptr.(*bool)
		case *string:
			check = *ptr.(*string) != ""
		case *time.Duration:
			check = *ptr.(*int64) > 0
		case *float64:
			check = *ptr.(*float64) != 0
		case *int:
			check = *ptr.(*int) != 0
		case *[]string:
			check = len(*ptr.(*[]string)) > 0
		case *os.File:
			check = *ptr.(**os.File) != nil
		default:
			// NOTE: Custom flags not supported.
			check = true
		}
		if !check {
			return fmt.Errorf("Missing --%s", name)
		}
	}
	return nil
}

type sliceValue struct {
	S []string
}

func (s *sliceValue) String() string {
	if s == nil || s.S == nil {
		return ""
	}
	return strings.Join(s.S, ",")
}

func (s *sliceValue) Set(value string) error {
	if value == "" {
		s.S = nil
	} else {
		s.S = strings.Split(value, ",")
	}
	return nil
}

type readfileValue struct {
	file *os.File
}

func (rf *readfileValue) String() string {
	if rf.file == nil {
		return ""
	}
	info, err := rf.file.Stat()
	if err != nil {
		return ""
	}
	return info.Name()
}

func (rf *readfileValue) Set(value string) error {
	var err error
	if value == "" {
		rf.file = nil
	} else {
		rf.file, err = os.Open(value)
	}
	return err
}

// String binds flag with string type.
func String(name string, defaultValue string, usage string) *string {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*string)
	}
	ptr := flag.String(name, defaultValue, usage)
	ptrs[name] = ptr
	return ptr
}

// Int binds flag with int type.
func Int(name string, defaultValue int, usage string) *int {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*int)
	}
	ptr := flag.Int(name, defaultValue, usage)
	ptrs[name] = ptr
	return ptr
}

// Bool binds flag with bool type.
func Bool(name string, defaultValue bool, usage string) *bool {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*bool)
	}
	ptr := flag.Bool(name, defaultValue, usage)
	ptrs[name] = ptr
	return ptr
}

// Float64 binds flag with float64 type.
func Float64(name string, defaultValue float64, usage string) *float64 {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*float64)
	}
	ptr := flag.Float64(name, defaultValue, usage)
	ptrs[name] = ptr
	return ptr
}

// Slice binds flag with slice type.
func Slice(name string, defaultValue []string, usage string) *[]string {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*[]string)
	}

	container := sliceValue{S: defaultValue}
	flag.Var(&container, name, usage)
	ptr := &container.S
	ptrs[name] = ptr
	return ptr
}

// Duration binds flag with time.Duration type.
func Duration(name string, defaultValue time.Duration, usage string) *time.Duration {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(*time.Duration)
	}
	ptr := flag.Duration(name, defaultValue, usage)
	ptrs[name] = ptr
	return ptr
}

// ReadFile binds flag with *os.File type. It will check the existent of file.
func ReadFile(name, path, usage string) **os.File {
	if ptr, exists := ptrs[name]; exists {
		return ptr.(**os.File)
	}
	container := readfileValue{}
	if path != "" {
		var err error
		container.file, err = os.Open(path)
		if err != nil {
			return nil
		}
	}
	flag.Var(&container, name, usage)
	ptr := &container.file
	ptrs[name] = ptr
	return ptr
}

type muxValue struct {
	values []flag.Value
}

func (cc *muxValue) String() string {
	if len(cc.values) > 0 {
		return cc.values[0].String()
	}
	return ""
}

func (cc *muxValue) Set(value string) error {
	for i, c := range cc.values {
		if err := c.Set(value); err != nil {
			return fmt.Errorf("Parse custom value: %d", i)
		}
	}
	return nil
}

func (cc *muxValue) Append(value flag.Value) {
	cc.values = append(cc.values, value)
}

// Var binds flag with custom value.
func Var(value flag.Value, name string, usage string) {
	if ptr, exists := ptrs[name]; exists {
		if cc, ok := ptr.(*muxValue); ok {
			cc.Append(value)
		}
		return
	}
	container := &muxValue{}
	container.Append(value)

	flag.Var(container, name, usage)

	ptr := container
	ptrs[name] = ptr
	return
}
