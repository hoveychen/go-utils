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
	project         bson.M
	client          *DbClient

	entryType reflect.Type
}

const defaultRefreshInterval = time.Minute * 5

type ReposOption func(*LocalRepos)

type ReposKVEntry struct {
	Key   string
	Value Hashable
}

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

func WithProjection(project bson.M) ReposOption {
	return func(repos *LocalRepos) {
		repos.project = project
	}
}

func WithClient(client *DbClient) ReposOption {
	return func(repos *LocalRepos) {
		repos.client = client
	}
}

type Hashable interface {
	GetId() string
}

func NewLocalRepos(db, col string, entryTmpl Hashable, opts ...ReposOption) *LocalRepos {
	repos := &LocalRepos{
		database:        db,
		collection:      col,
		client:          getClient(db, col),
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
	c, s := r.client.Open(r.database, r.collection)
	defer s.Close()

	newData := map[string]Hashable{}
	find := c.Find(r.query)
	if r.project != nil {
		find = find.Select(r.project)
	}
	iter := find.Iter()
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

func (r *LocalRepos) AllValues() []Hashable {
	r.RLock()
	defer r.RUnlock()
	var ret []Hashable
	for _, v := range r.data {
		ret = append(ret, v)
	}
	return ret
}

func (r *LocalRepos) AllKeys() []string {
	r.RLock()
	defer r.RUnlock()
	var ret []string
	for k := range r.data {
		ret = append(ret, k)
	}
	return ret
}

func (r *LocalRepos) AllItems() []*ReposKVEntry {
	r.RLock()
	defer r.RUnlock()
	var ret []*ReposKVEntry
	for k, v := range r.data {
		ret = append(ret, &ReposKVEntry{k, v})
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
