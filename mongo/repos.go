package mongo

import (
	"reflect"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
)

// LocalRepos is a Key/Value map that cached the data from a particular mongodb
// to local memory, and keep refreshing data in certian intervals.
// It's useful for small collection with high read I/O.
type LocalRepos struct {
	sync.RWMutex
	database        string
	collection      string
	refreshInterval time.Duration
	data            map[string]Hashable
	ticker          *time.Ticker
	query           bson.M

	entryType reflect.Type
}

const defaultRefreshInterval = time.Minute * 5

type ReposOption func(*LocalRepos)

func WithRefreshInterval(dur time.Duration) ReposOption {
	return func(repos *LocalRepos) {
		repos.refreshInterval = dur
	}
}

func WithQuery(query bson.M) ReposOption {
	return func(repos *LocalRepos) {
		repos.query = query
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

	r.ticker = time.NewTicker(r.refreshInterval)
	go func() {
		for range r.ticker.C {
			r.reloadEntries()
		}
	}()
}

func (r *LocalRepos) reloadEntries() error {
	c, s := Open(r.database, r.collection)
	defer s.Close()

	newData := map[string]Hashable{}
	iter := c.Find(r.query).Iter()
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

	r.Lock()
	defer r.Unlock()

	r.data = newData
	return nil
}

func (r *LocalRepos) All() []Hashable {
	r.RLock()
	defer r.RUnlock()
	var ret []Hashable
	for _, item := range r.data {
		ret = append(ret, item)
	}
	return ret
}

func (r *LocalRepos) Get(id string) Hashable {
	r.RLock()
	defer r.RUnlock()
	return r.data[id]
}

func (r *LocalRepos) Len() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.data)
}

func (r *LocalRepos) Close() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
}
