// Package cache provides MemCache as a TTL based simple cache in memory.
package cache

import (
	"github.com/hoveychen/go-utils"
	"sync"
	"time"
)

type MemCache struct {
	Items  map[string]*cachedItem
	lock   sync.RWMutex
	ticker *time.Ticker
}

type cachedItem struct {
	Payload    interface{}
	ExpireTime time.Time
}

func NewMemCache() *MemCache {
	c := &MemCache{
		Items: map[string]*cachedItem{},
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			c.checkExpire()
		}
	}()
	return c
}

func (c *MemCache) checkExpire() {
	c.lock.Lock()
	defer c.lock.Unlock()
	// TODO(Yuheng): Optimize the expire checking algorithm.
	now := goutils.GetNow()
	for key, item := range c.Items {
		if item.ExpireTime.Before(now) {
			delete(c.Items, key)
		}
	}
}

func (c *MemCache) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	i, hit := c.Items[key]
	if hit {
		return i.Payload
	} else {
		return nil
	}
}

func (c *MemCache) UpsertWithTTL(key string, val interface{}, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	i, hit := c.Items[key]
	if hit {
		i.Payload = val
		i.ExpireTime = goutils.GetNow().Add(ttl)
	} else {
		c.Items[key] = &cachedItem{
			Payload:    val,
			ExpireTime: goutils.GetNow().Add(ttl),
		}
	}
}

func (c *MemCache) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ticker.Stop()
	c.Items = nil
}
