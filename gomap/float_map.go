package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// FloatMap is a Map with values as built-in float64 type.
type FloatMap struct {
	sync.RWMutex
	data map[string]float64
}

type FloatMapEntry struct {
	Key   string
	Value float64
}

func NewFloatMap() *FloatMap {
	return &FloatMap{
		data: map[string]float64{},
	}
}

func WrapFloatMap(m map[string]float64) *FloatMap {
	if m == nil {
		return NewFloatMap()
	}
	return &FloatMap{
		data: m,
	}
}

func (m *FloatMap) Unwrap() map[string]float64 {
	m.Lock()
	defer m.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *FloatMap) Clone() *FloatMap {
	m.RLock()
	defer m.RUnlock()

	newData := map[string]float64{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapFloatMap(newData)
}

func (m *FloatMap) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

func (m *FloatMap) UnmarshalJSON(d []byte) error {
	m.Lock()
	defer m.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *FloatMap) Set(key string, value float64) {
	m.Lock()
	m.data[key] = value
	m.Unlock()
}

func (m *FloatMap) Delete(key string) {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
}

func (m *FloatMap) Get(key string) float64 {
	m.RLock()
	defer m.RUnlock()
	return m.data[key]
}

func (m *FloatMap) Exists(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *FloatMap) GetKeysUnordered() []string {
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

func (m *FloatMap) GetKeys() []string {
	keys := m.GetKeysUnordered()
	sort.Strings(keys)
	return keys
}

func (m *FloatMap) GetValues() []float64 {
	m.RLock()
	defer m.RUnlock()
	ret := make([]float64, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

func (m *FloatMap) GetItemsUnordered() []FloatMapEntry {
	m.RLock()
	defer m.RUnlock()

	ret := make([]FloatMapEntry, len(m.data))
	i := 0
	for k, v := range m.data {
		ret[i] = FloatMapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return ret
}

func (m *FloatMap) GetItems() []FloatMapEntry {
	m.RLock()
	defer m.RUnlock()

	keys := m.GetKeys()
	ret := make([]FloatMapEntry, len(keys))
	i := 0
	for _, k := range keys {
		ret[i] = FloatMapEntry{
			Key:   k,
			Value: m.data[k],
		}
		i++
	}

	return ret
}

func (m *FloatMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.data)
}
