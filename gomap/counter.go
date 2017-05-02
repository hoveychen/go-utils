package gomap

import "sync"

import "sort"

// Counter provides comprehensive statistic functions over a bunch of numbers.
// Like avg, sum, peek-trimming dataset.
// TODO(yuheng): Implmenet data compacting when memory grows to unignorable size.
type Counter struct {
	data map[int]int
	lock sync.RWMutex
}

func NewCounter() *Counter {
	return &Counter{
		data: map[int]int{},
	}
}

func (c *Counter) Add(value ...int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, v := range value {
		c.data[v]++
	}
}

func (c *Counter) AddMultipleTimes(value, times int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data[value] += times
}

func (c *Counter) Mean() int {
	n := c.Len()
	if n == 0 {
		return 0
	}
	mid := int(n / 2)
	pos := 0
	for _, item := range c.Detail() {
		l := pos
		r := pos + item.Freq
		if l <= mid && mid < r {
			return item.Value
		}
		pos += item.Freq
	}
	return 0
}

func (c *Counter) Sum() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	sum := 0
	for v, num := range c.data {
		sum += v * num
	}
	return sum
}

func (c *Counter) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	n := 0
	for _, num := range c.data {
		n += num
	}
	return n
}

func (c *Counter) Avg() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	n := c.Len()
	if n == 0 {
		return 0.0
	}
	return float64(c.Sum()) / float64(n)
}

// Deviation returns the population standard deviation of numbers.
func (c *Counter) Deviation() float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	n := c.Len()
	if n == 0 {
		return 0.0
	}
	var sqrSum int64
	for v, num := range c.data {
		sqrSum += int64(v) * int64(v) * int64(num)
	}
	avg := c.Avg()
	return float64(sqrSum)/float64(n) - avg*avg
}

type FrequencyEntry struct {
	Value int
	Freq  int
}

func (c *Counter) Detail() []*FrequencyEntry {
	values := make([]*FrequencyEntry, len(c.data))

	cnt := 0
	for v, f := range c.data {
		values[cnt] = &FrequencyEntry{
			Value: v,
			Freq:  f,
		}
		cnt++
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Value < values[j].Value
	})
	return values
}

// TrimTop returns another counter containing partical elements
// within [leastRatio, mostRatio] part in order from smallest to
// largest.
func (c *Counter) TrimTop(leastRatio, mostRatio float64) *Counter {
	values := c.Detail()

	ret := NewCounter()
	n := c.Len()
	if leastRatio < 0 {
		leastRatio = 0
	}
	if mostRatio > 1 {
		mostRatio = 1
	}
	start := int(float64(n) * leastRatio)
	end := int(float64(n) * mostRatio)
	pos := 0
	// Alternative form is to get the flatten value slice [start:end]
	for _, entry := range values {
		l := pos
		r := pos + entry.Freq

		if l < start {
			l = start
		}
		if r > end {
			r = end
		}
		if r-l > 0 {
			ret.AddMultipleTimes(entry.Value, r-l)
		}
		pos += entry.Freq
	}
	return ret
}
