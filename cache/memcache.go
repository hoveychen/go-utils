// Package cache provides MemCache as a TTL based simple cache in memory.
package cache

import (
	"sync"
	"time"

	"github.com/hoveychen/go-utils"
)

type MemCache struct {
	Items  map[string]*cachedItem
	lock   sync.RWMutex
	ticker *time.Ticker
}

type cachedItem struct {
	Payload    interface{}
	ExpireTime time.Time
	TTL        time.Duration
}

func NewMemCache(checkInterval time.Duration) *MemCache {
	c := &MemCache{
		Items: map[string]*cachedItem{},
	}
	go func() {
		ticker := time.NewTicker(checkInterval)
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
		// It's hit. Extends the expiring time by another TTL.
		i.ExpireTime = goutils.GetNow().Add(i.TTL)
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
		i.TTL = ttl
		i.ExpireTime = goutils.GetNow().Add(ttl)
	} else {
		c.Items[key] = &cachedItem{
			Payload:    val,
			ExpireTime: goutils.GetNow().Add(ttl),
			TTL:        ttl,
		}
	}
}

func (c *MemCache) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ticker.Stop()
	c.Items = nil
}
