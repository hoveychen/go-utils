package sort

import (
	"testing"
)

type Class struct {
	v int
}

func (c *Class) Rank() float64 {
	return float64(c.v)
}

func TestSort(t *testing.T) {
	input := []*Class{
		{2}, {1}, {3},
	}
	output := []*Class{
		{1}, {2}, {3},
	}

	err := Sort(input)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < len(input); i++ {
		if input[i].v != output[i].v {
			t.Error("Incorrect sorting, actual", input, "expect", output)
			break
		}
	}
}

func TestSortPtr(t *testing.T) {
	input := []*Class{
		{2}, {1}, {3},
	}
	output := []*Class{
		{1}, {2}, {3},
	}

	err := Sort(&input)
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < len(input); i++ {
		if input[i].v != output[i].v {
			t.Error("Incorrect sorting, actual", input, "expect", output)
			break
		}
	}
}
