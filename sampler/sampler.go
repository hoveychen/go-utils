// This package provide a simple way to sample steam data.
package sampler

import "math/rand"

type Sampler struct {
	data  []interface{}
	size  int
	cap   int
	total int
}

func New(cap int) *Sampler {
	return &Sampler{
		data:  make([]interface{}, cap),
		size:  0,
		total: 0,
		cap:   cap,
	}
}

func (s *Sampler) Feed(obj interface{}) {
	s.total++
	if s.size < s.cap {
		s.data[s.size] = obj
		s.size++
	} else if rand.Float64() <= float64(s.cap)/float64(s.total) {
		idx := rand.Int() % s.cap
		s.data[idx] = obj
	}
}

func (s *Sampler) Total() int {
	return s.total
}

func (s *Sampler) Cap() int {
	return s.cap
}

func (s *Sampler) Output() []interface{} {
	return s.data[:s.size]
}
