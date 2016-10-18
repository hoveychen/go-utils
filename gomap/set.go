package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// Set is a Map with values as built-in bool type.
type Set struct {
	data map[string]bool
	lock sync.RWMutex
}

func NewSet() *Set {
	return &Set{
		data: map[string]bool{},
	}
}

func (m *Set) removeInvalid() {
	// A book keeper to make sure no invalid (false value) values leaks to outside.
	for k, v := range m.data {
		if !v {
			delete(m.data, k)
		}
	}
}

func WrapSet(m map[string]bool) *Set {
	if m == nil {
		return NewSet()
	}
	s := &Set{
		data: m,
	}
	s.removeInvalid()
	return s
}

func (m *Set) Unwrap() map[string]bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *Set) Clone() *Set {
	m.lock.RLock()
	defer m.lock.RUnlock()

	newData := map[string]bool{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapSet(newData)
}

func (m *Set) MarshalJSON() ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	d, err := json.Marshal(m.data)

	return d, err
}

func (m *Set) UnmarshalJSON(d []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if err := json.Unmarshal(d, &m.data); err != nil {
		return err
	} else {
		m.removeInvalid()
		return nil
	}
}

func (m *Set) Add(elem string) {
	m.lock.Lock()
	m.data[elem] = true
	m.lock.Unlock()
}

func (m *Set) Remove(elem string) {
	m.lock.Lock()
	delete(m.data, elem)
	m.lock.Unlock()
}

func (m *Set) Contains(elem string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[elem]
}

func (m *Set) GetElementsUnordered() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := []string{}
	for k, v := range m.data {
		if v {
			ret = append(ret, k)
		}
	}
	return ret
}

func (m *Set) GetElements() []string {
	elems := m.GetElementsUnordered()
	sort.Strings(elems)
	return elems
}

func (m *Set) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}
