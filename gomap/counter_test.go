package gomap

import (
	"math"
	"testing"
)

func TestEmptyCounter(t *testing.T) {
	c := NewCounter()
	if c.Len() != 0 || c.Sum() != 0 || c.Avg() != 0 || c.Deviation() != 0 {
		t.Error("Non-empty counter when initialized.")
	}
	result := c.TrimTop(0.2, 0.8)
	if result.Len() != 0 || result.Sum() != 0 || result.Avg() != 0 || result.Deviation() != 0 {
		t.Error("Non-empty counter after trimmed.")
	}
}

func TestCounterDetail(t *testing.T) {
	c := NewCounter()
	c.Add(1, 2, 3, 4, 1, 6, 7, 6, 9, 10, 1)
	expVal := []int{1, 2, 3, 4, 6, 7, 9, 10}
	expFreq := []int{3, 1, 1, 1, 2, 1, 1, 1}
	d := c.Detail()
	if len(d) != len(expVal) {
		t.Error("Expected Value Len:", len(expVal), "Actual", len(d))
	}

	for i, v := range d {
		if expVal[i] != v.Value {
			t.Error("Pos:", i, "Expected value:", expVal[i], "Actual", v.Value)
		}
		if expFreq[i] != v.Freq {
			t.Error("Pos:", i, "Expected value:", expFreq[i], "Actual", v.Freq)
		}
	}

}

func TestTrimCounter(t *testing.T) {
	c := NewCounter()
	c.Add(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	expVal := []int{4, 5, 6}
	result := c.TrimTop(0.3, 0.6)
	if result.Len() != len(expVal) {
		t.Error("Error trimmed result len.", result.Len(), "Expected", len(expVal))
	}
	d := result.Detail()
	if len(d) != len(expVal) {
		t.Error("Error trimmed detail len.")
	}
	for i, v := range expVal {
		if d[i].Value != v {
			t.Error("Error trimmed detail value.")
		}
		if d[i].Freq != 1 {
			t.Error("Error trimmed detail frequency.")
		}
	}
}

func TestCounterStatFunc(t *testing.T) {
	c := NewCounter()
	c.Add(10, 2, 38, 23, 38, 23, 21)

	expLen := 7
	if c.Len() != expLen {
		t.Error("Error sum", c.Len(), "Expected", expLen)
	}

	expDeviation := 151.2653
	if math.Abs(c.Deviation()-expDeviation) > 1e-4 {
		t.Error("Error deviation.", c.Deviation(), "Expected", expDeviation)
	}

	expSum := 155
	if c.Sum() != expSum {
		t.Error("Error sum", c.Sum(), "Expected", expSum)
	}

	expAvg := 22.1428
	if math.Abs(c.Avg()-expAvg) > 1e-4 {
		t.Error("Error sum", c.Avg(), "Expected", expAvg)
	}

}
