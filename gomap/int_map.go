package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// IntMap is a Map with values as built-in int type.
type IntMap struct {
	data map[string]int
	lock sync.RWMutex
}

type IntMapEntry struct {
	Key   string
	Value int
}

func NewIntMap() *IntMap {
	return &IntMap{
		data: map[string]int{},
	}
}

func WrapIntMap(m map[string]int) *IntMap {
	if m == nil {
		return NewIntMap()
	}
	return &IntMap{
		data: m,
	}
}

func (m *IntMap) Unwrap() map[string]int {
	m.lock.Lock()
	defer m.lock.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *IntMap) Clone() *IntMap {
	m.lock.RLock()
	defer m.lock.RUnlock()

	newData := map[string]int{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapIntMap(newData)
}

func (m *IntMap) MarshalJSON() ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

func (m *IntMap) UnmarshalJSON(d []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *IntMap) Set(key string, value int) {
	m.lock.Lock()
	m.data[key] = value
	m.lock.Unlock()
}

func (m *IntMap) Delete(key string) {
	m.lock.Lock()
	delete(m.data, key)
	m.lock.Unlock()
}

func (m *IntMap) Get(key string) int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[key]
}

func (m *IntMap) Exists(key string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *IntMap) GetKeysUnordered() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]string, len(m.data))
	i := 0
	for k, _ := range m.data {
		ret[i] = k
		i++
	}

	return ret
}

func (m *IntMap) GetKeys() []string {
	keys := m.GetKeysUnordered()
	sort.Strings(keys)
	return keys
}

func (m *IntMap) GetValues() []int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret := make([]int, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

func (m *IntMap) GetItemsUnordered() []IntMapEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ret := make([]IntMapEntry, len(m.data))
	i := 0
	for k, v := range m.data {
		ret[i] = IntMapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return ret
}

func (m *IntMap) GetItems() []IntMapEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()

	keys := m.GetKeys()
	ret := make([]IntMapEntry, len(keys))
	i := 0
	for _, k := range keys {
		ret[i] = IntMapEntry{
			Key:   k,
			Value: m.data[k],
		}
		i++
	}

	return ret
}

func (m *IntMap) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}
