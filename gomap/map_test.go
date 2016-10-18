package gomap

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func TestWrappingEmpty(t *testing.T) {
	m := Wrap(nil)
	if m.Len() != 0 {
		t.Error("Wrap nil should be the same as empty.")
	}
}

func TestWrapping(t *testing.T) {
	contact := map[string]interface{}{
		"key1": "hello",
		"key2": "world",
	}

	m := Wrap(contact)
	if m.Len() != 2 || m.Get("key1").(string) != "hello" || m.Get("key2").(string) != "world" {
		t.Error("Unexpected content in the wrapping map.")
	}
}

func newSampleMap() *Map {
	return Wrap(map[string]interface{}{
		"key1": 1,
		"key3": 2,
		"key2": 3,
	})
}

func TestClone(t *testing.T) {
	m := newSampleMap()
	newM := m.Clone()
	if newM.Len() != m.Len() {
		t.Error("Clone results different size")
	}

	for _, i := range m.GetItems() {
		if newM.Get(i.Key).(int) != i.Value.(int) {
			t.Error("Clone results different content")
		}
	}

	newM.Set("key2", 100)
	if m.Get("key2").(int) == 100 {
		t.Error("Clone refers to the old memory")
	}
}

func TestGetAndSet(t *testing.T) {
	m := newSampleMap()
	if m.Len() != 3 {
		t.Error("Length not correct")
	}
	if m.Get("key2").(int) != 3 {
		t.Error("Get not correct")
	}
	if m.Get("something") != nil || m.Get("") != nil {
		t.Error("Get returns unexpected content")
	}
	if m.Exists("key4") {
		t.Error("Exists return non-exists result.")
	}
	m.Set("key4", 4)
	if m.Len() != 4 {
		t.Error("Length don't grow after new element")
	}
	if !m.Exists("key4") {
		t.Error("Exists return exists result as false.")
	}
	if m.Get("key4").(int) != 4 {
		t.Error("Set not taking effect")
	}
	m.Delete("key1")
	if m.Len() != 3 {
		t.Error("Length don't shrink after delete element")
	}
	if m.Get("key1") != nil {
		t.Error("Delete not taking effect")
	}
	if m.Exists("key1") {
		t.Error("Exists return non-exists result.")
	}
}

func TestConcurrency(t *testing.T) {
	// If no seg error, it should be thread-safe??
	m := New()
	wg := sync.WaitGroup{}
	wg.Add(100 * 6)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprint(rand.Int() % 100)
				m.Set(key, j)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprint(rand.Int() % 100)
				m.Get(key)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := fmt.Sprint(rand.Int() % 100)
				m.Delete(key)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.GetKeysUnordered()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.GetValues()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				m.GetItemsUnordered()
			}
		}()
	}
	wg.Wait()
}

func TestGetKeys(t *testing.T) {
	m := newSampleMap()
	keys := m.GetKeys()
	for i := 1; i <= 3; i++ {
		if keys[i-1] != fmt.Sprintf("key%d", i) {
			t.Error("GetKeys not returning in order")
		}
	}
}

func TestGetItems(t *testing.T) {
	m := newSampleMap()
	items := m.GetItems()
	expectedVal := []int{1, 3, 2}
	for i := 1; i <= 3; i++ {
		if items[i-1].Key != fmt.Sprintf("key%d", i) || items[i-1].Value.(int) != expectedVal[i-1] {
			t.Error("GetItems not returning in order")
		}
	}
}
