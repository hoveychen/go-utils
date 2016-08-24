// Package sort provides ranking method to easily sort a slice.
// Golang's stock sorting package is nice, but not easy to use.
// This package define another way to easily sort a slice inplace, all user need
// is to implement the Ranker interface, aka. a Rank() method to the element type in the slice.
// Example usage:
//
// type Medals struct {
//     Gold, Silver, Bronze int
// }
//
// func (m *Modals) Rank() float64 {
//     return float64(m.Gold * 10000 + m.Silver * 100 + m.Bronze)
// }
//
// medals := []*Medals{
//    {1,3,3},
//    {4,2,1},
//    {1,2,3},
// }
// err := sort.Sort(medals)
// if err != nil {
//     ....
// }
package sort

import (
	"errors"
	"reflect"
	gosort "sort"
)

type Ranker interface {
	Rank() float64
}

type rankSorter struct {
	order []int
	rank  []float64
}

func (s *rankSorter) Less(i, j int) bool {
	return s.rank[s.order[i]] < s.rank[s.order[j]]
}

func (s *rankSorter) Swap(i, j int) {
	s.order[i], s.order[j] = s.order[j], s.order[i]
}

func (s *rankSorter) Len() int {
	return len(s.order)
}

func newRankSorter(list []Ranker) *rankSorter {
	sorter := &rankSorter{}
	for i, item := range list {
		sorter.order = append(sorter.order, i)
		sorter.rank = append(sorter.rank, item.Rank())
	}
	return sorter
}

func convertRankerSlice(slicePtr interface{}) ([]Ranker, error) {
	s := reflect.ValueOf(slicePtr)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Slice {
		return nil, errors.New("Pointed element is not a slice.")
	}
	ret := make([]Ranker, s.Len())
	for i := 0; i < s.Len(); i++ {
		r, ok := s.Index(i).Interface().(Ranker)
		if !ok {
			return nil, errors.New("Item doesn't implement Ranker interface.")
		}
		ret[i] = r
	}

	return ret, nil
}

func sortByFunc(slicePtr interface{}, f func(sorter *rankSorter)) error {
	rankerSlice, err := convertRankerSlice(slicePtr)
	if err != nil {
		return err
	}
	sorter := newRankSorter(rankerSlice)
	f(sorter)

	v := reflect.ValueOf(slicePtr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i, j := range sorter.order {
		v.Index(i).Set(reflect.ValueOf(rankerSlice[j]))
	}

	return nil
}

func Sort(slicePtr interface{}) error {
	return sortByFunc(slicePtr, func(sorter *rankSorter) {
		gosort.Sort(sorter)
	})
}

func ReverseSort(slicePtr interface{}) error {
	return sortByFunc(slicePtr, func(sorter *rankSorter) {
		gosort.Sort(gosort.Reverse(sorter))
	})
}

func StableSort(slicePtr interface{}) error {
	return sortByFunc(slicePtr, func(sorter *rankSorter) {
		gosort.Stable(sorter)
	})
}

func ReverseStableSort(slicePtr interface{}) error {
	return sortByFunc(slicePtr, func(sorter *rankSorter) {
		gosort.Stable(gosort.Reverse(sorter))
	})
}
