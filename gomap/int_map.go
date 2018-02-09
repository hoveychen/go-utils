package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// IntMap is a Map with values as built-in int type.
type IntMap struct {
	sync.RWMutex
	data map[string]int
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
	m.Lock()
	defer m.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *IntMap) Clone() *IntMap {
	m.RLock()
	defer m.RUnlock()

	newData := map[string]int{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapIntMap(newData)
}

func (m *IntMap) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

func (m *IntMap) UnmarshalJSON(d []byte) error {
	m.Lock()
	defer m.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *IntMap) Set(key string, value int) {
	m.Lock()
	m.data[key] = value
	m.Unlock()
}

func (m *IntMap) Add(key string, value int) {
	m.Lock()
	if _, hit := m.data[key]; hit {
		m.data[key] += value
	} else {
		m.data[key] = value
	}
	m.Unlock()
}

func (m *IntMap) Delete(key string) {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
}

func (m *IntMap) Get(key string) int {
	m.RLock()
	defer m.RUnlock()
	return m.data[key]
}

func (m *IntMap) Exists(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *IntMap) GetKeysUnordered() []string {
	m.RLock()
	defer m.RUnlock()

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
	m.RLock()
	defer m.RUnlock()
	ret := make([]int, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

// GetTopN returns top N entries in desc value order.
func (m *IntMap) GetTopN(n int) []IntMapEntry {
	items := m.GetItemsUnordered()
	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	if n > len(items) {
		n = len(items)
	}
	return items[:n]
}

func (m *IntMap) GetItemsUnordered() []IntMapEntry {
	m.RLock()
	defer m.RUnlock()

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
	m.RLock()
	defer m.RUnlock()

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
	m.RLock()
	defer m.RUnlock()
	return len(m.data)
}
