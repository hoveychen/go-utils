package goutils

import (
	"testing"
)

func TestTimezone(t *testing.T) {
	if ChinaTimezone == nil {
		t.Fatal("Failed to parse timezone")
	}
}
