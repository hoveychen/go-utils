package mongo

import (
	"reflect"
	"sync"
	"time"
)

// LocalRepos is a Key/Value map that cached the data from a particular mongodb
// to local memory, and keep refreshing data in certian intervals.
// It's useful for small collection with high read I/O.
type LocalRepos struct {
	database        string
	collection      string
	refreshInterval time.Duration
	refreshing      sync.RWMutex
	data            map[string]Hashable

	entryType reflect.Type
}

const defaultRefreshInterval = time.Minute * 5

type ReposOption func(*LocalRepos)

func WithRefreshInterval(dur time.Duration) ReposOption {
	return func(repos *LocalRepos) {
		repos.refreshInterval = dur
	}
}

type Hashable interface {
	GetId() string
}

func NewLocalRepos(db, col string, entryTmpl Hashable, opts ...ReposOption) *LocalRepos {
	repos := &LocalRepos{
		database:        db,
		collection:      col,
		refreshInterval: defaultRefreshInterval,
	}
	for _, opt := range opts {
		opt(repos)
	}

	if reflect.TypeOf(entryTmpl).Kind() == reflect.Ptr {
		repos.entryType = reflect.ValueOf(entryTmpl).Elem().Type()
	} else {
		repos.entryType = reflect.ValueOf(entryTmpl).Type()
	}
	return repos
}

func (r *LocalRepos) Init() {
	r.reloadEntries()

	go func() {
		for range time.Tick(r.refreshInterval) {
			r.reloadEntries()
		}
	}()
}

func (r *LocalRepos) reloadEntries() error {
	c, s := Open(r.database, r.collection)
	defer s.Close()

	newData := map[string]Hashable{}
	iter := c.Find(nil).Iter()
	for {
		newVal := reflect.New(r.entryType).Interface().(Hashable)
		if !iter.Next(newVal) {
			break
		}
		id := newVal.GetId()
		newData[id] = newVal
	}

	if iter.Err() != nil {
		return iter.Err()
	}

	r.refreshing.Lock()
	defer r.refreshing.Unlock()

	r.data = newData
	return nil
}

func (r *LocalRepos) Get(id string) Hashable {
	r.refreshing.RLock()
	defer r.refreshing.RUnlock()
	return r.data[id]
}

func (r *LocalRepos) Len() int {
	r.refreshing.RLock()
	defer r.refreshing.RUnlock()
	return len(r.data)
}
