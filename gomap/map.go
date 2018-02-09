package gomap

import (
	"encoding/json"
	"sort"
	"sync"
)

// Map is a simple implementation of thread-safe key-value structure just like the built-in map.
// It may not be quite efficient, but most of method should meet daily need for multi-goroutine
// environment.
type Map struct {
	sync.RWMutex
	data map[string]interface{}
}

// MapEntry is the return structure for iterating.
type MapEntry struct {
	Key   string
	Value interface{}
}

// New creates a new map structure.
func New() *Map {
	return &Map{
		data: map[string]interface{}{},
	}
}

// Wrap takes the ownership of a built-in map, and return a new map structure.
func Wrap(m map[string]interface{}) *Map {
	if m == nil {
		return New()
	}
	return &Map{
		data: m,
	}
}

// Unwrap releases the ownership of the inner built-in map, and return it.
// Note that the map structure also remove the reference to this map object,
// which means this map structure won't function any more but in the ease of
// the risk of memory leak.
func (m *Map) Unwrap() map[string]interface{} {
	m.Lock()
	defer m.Unlock()

	d := m.data
	m.data = nil
	return d
}

// Clone shallow copies the keys and values to a new map structure.
func (m *Map) Clone() *Map {
	m.RLock()
	defer m.RUnlock()

	newData := map[string]interface{}{}
	for k, v := range m.data {
		newData[k] = v
	}

	return Wrap(newData)
}

// MarshalJSON implements the json.Marshaller interface.
func (m *Map) MarshalJSON() ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	d, err := json.Marshal(m.data)
	return d, err
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (m *Map) UnmarshalJSON(d []byte) error {
	m.Lock()
	defer m.Unlock()

	return json.Unmarshal(d, &m.data)
}

func (m *Map) Set(key string, value interface{}) {
	m.Lock()
	m.data[key] = value
	m.Unlock()
}

func (m *Map) Delete(key string) {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
}

func (m *Map) Get(key string) interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.data[key]
}

func (m *Map) Exists(key string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok
}

// GetOrCreate gets the value by key. If no value exists, it will call the createFn() to generate a new object.
// It's useful to implement a singleton cache flow, where you only
// want to create each key/value exactly once.
// IMPORTANT NOTE: The createFn should not invoke any method in this
// map, otherwise it will DEADLOCK.
func (m *Map) GetOrCreate(key string, createFn func() interface{}) interface{} {
	m.RLock()
	_, ok := m.data[key]
	if ok {
		defer m.RUnlock()
		return m.data[key]
	} else {
		// Call createFn to generate the new object set to key.
		m.RUnlock()
		m.Lock()
		defer m.Unlock()
		value := createFn()
		m.data[key] = value
		return value
	}
}

// GetKeysUnordered returns all the *copy* of keys, but no order is guaranteed.
func (m *Map) GetKeysUnordered() []string {
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

// GetKeys returns all the *copy* of keys, and sort them in alphabet order.
func (m *Map) GetKeys() []string {
	keys := m.GetKeysUnordered()
	sort.Strings(keys)
	return keys
}

// GetValues returns all the *copy* of values, but no order is guaranteed.
func (m *Map) GetValues() []interface{} {
	m.RLock()
	defer m.RUnlock()
	ret := make([]interface{}, len(m.data))
	i := 0
	for _, v := range m.data {
		ret[i] = v
		i++
	}

	return ret
}

// GetItemsUnordered returns all the *copy* of key/value pairs, but no order is guaranteed.
func (m *Map) GetItemsUnordered() []MapEntry {
	m.RLock()
	defer m.RUnlock()

	ret := make([]MapEntry, len(m.data))
	i := 0
	for k, v := range m.data {
		ret[i] = MapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return ret
}

// GetItems returns all the *copy* of key/value pairs, and sort them by keys in alphabet order.
func (m *Map) GetItems() []MapEntry {
	m.RLock()
	defer m.RUnlock()

	keys := m.GetKeys()
	ret := make([]MapEntry, len(keys))
	i := 0
	for _, k := range keys {
		ret[i] = MapEntry{
			Key:   k,
			Value: m.data[k],
		}
		i++
	}

	return ret
}

// Len returns the size of the map.
func (m *Map) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.data)
}
