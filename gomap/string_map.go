package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// StringMap is a Map with values as built-in string type.
type StringMap struct {
	sync.RWMutex
	data map[string]string
}

type StringMapEntry struct {
	Key   string
	Value string
}

func NewStringMap() *StringMap {
	return &StringMap{
		data: map[string]string{},
	}
}

func WrapStringMap(m map[string]string) *StringMap {
	if m == nil {
		return NewStringMap()
	}
	return &StringMap{
		data: m,
	}
}

func (m *StringMap) Unwrap() map[string]string {
	m.Lock()
	defer m.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *StringMap) Clone() *StringMap {
	m.RLock()
	defer m.RUnlock()

	newData := map[string]string{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapStringMap(newData)
}

func (m *StringMap) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

func (m *StringMap) UnmarshalJSON(d []byte) error {
	m.Lock()
	defer m.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *StringMap) Set(key string, value string) {
	m.Lock()
	m.data[key] = value
	m.Unlock()
}

func (m *StringMap) Delete(key string) {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
}

func (m *StringMap) Get(key string) string {
	m.RLock()
	defer m.RUnlock()
	return m.data[key]
}

func (m *StringMap) Exists(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *StringMap) GetKeysUnordered() []string {
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

func (m *StringMap) GetKeys() []string {
	keys := m.GetKeysUnordered()
	sort.Strings(keys)
	return keys
}

func (m *StringMap) GetValues() []string {
	m.RLock()
	defer m.RUnlock()
	ret := make([]string, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

func (m *StringMap) GetItemsUnordered() []StringMapEntry {
	m.RLock()
	defer m.RUnlock()

	ret := make([]StringMapEntry, len(m.data))
	i := 0
	for k, v := range m.data {
		ret[i] = StringMapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return ret
}

func (m *StringMap) GetItems() []StringMapEntry {
	m.RLock()
	defer m.RUnlock()

	keys := m.GetKeys()
	ret := make([]StringMapEntry, len(keys))
	i := 0
	for _, k := range keys {
		ret[i] = StringMapEntry{
			Key:   k,
			Value: m.data[k],
		}
		i++
	}

	return ret
}

func (m *StringMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.data)
}
