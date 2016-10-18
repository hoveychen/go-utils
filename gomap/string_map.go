package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// StringMap is a Map with values as built-in string type.
type StringMap struct {
	data map[string]string
	lock sync.RWMutex
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
	m.lock.Lock()
	defer m.lock.Unlock()

	d := m.data
	m.data = nil
	return d
}

func (m *StringMap) Clone() *StringMap {
	m.lock.RLock()
	defer m.lock.RUnlock()

	newData := map[string]string{}
	for k, v := range m.data {
		newData[k] = v
	}

	return WrapStringMap(newData)
}

func (m *StringMap) MarshalJSON() ([]byte, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

func (m *StringMap) UnmarshalJSON(d []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *StringMap) Set(key string, value string) {
	m.lock.Lock()
	m.data[key] = value
	m.lock.Unlock()
}

func (m *StringMap) Delete(key string) {
	m.lock.Lock()
	delete(m.data, key)
	m.lock.Unlock()
}

func (m *StringMap) Get(key string) string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.data[key]
}

func (m *StringMap) Exists(key string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *StringMap) GetKeysUnordered() []string {
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

func (m *StringMap) GetKeys() []string {
	keys := m.GetKeysUnordered()
	sort.Strings(keys)
	return keys
}

func (m *StringMap) GetValues() []string {
	m.lock.RLock()
	defer m.lock.RUnlock()
	ret := make([]string, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

func (m *StringMap) GetItemsUnordered() []StringMapEntry {
	m.lock.RLock()
	defer m.lock.RUnlock()

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
	m.lock.RLock()
	defer m.lock.RUnlock()

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
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.data)
}
