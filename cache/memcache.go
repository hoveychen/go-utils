// Package cache provides MemCache as a TTL based simple cache in memory.
package cache

import (
	"time"

	"github.com/hoveychen/go-utils"
	"github.com/hoveychen/go-utils/gomap"
)

type MemCache struct {
	items  *gomap.Map
	ticker *time.Ticker
}

type cachedItem struct {
	Payload    interface{}
	ExpireTime time.Time
	TTL        time.Duration
}

func NewMemCache(recycleInterval time.Duration) *MemCache {
	c := &MemCache{
		items: gomap.New(),
	}

	go func() {
		c.ticker = time.NewTicker(recycleInterval)
		for range c.ticker.C {
			c.removeExpired()
		}
	}()
	return c
}

func (c *MemCache) removeExpired() {
	now := goutils.GetNow()
	for _, result := range c.items.GetItemsUnordered() {
		item := result.Value.(*cachedItem)
		if item == nil {
			goutils.LogError("Unexpected values in memcache")
			continue
		}
		if item.ExpireTime.Before(now) {
			c.items.Delete(result.Key)
		}
	}
}

func (c *MemCache) Get(key string) interface{} {
	i := c.items.Get(key)

	if i == nil {
		return nil
	} else {
		item := i.(*cachedItem)
		if item == nil {
			goutils.LogError("Unexpected values in memcache")
			return nil
		}
		if item.ExpireTime.Before(goutils.GetNow()) {
			// This item had expired.
			c.items.Delete(key)
			return nil
		}
		// It's hit. Extends the expiring time by another TTL.
		item.ExpireTime = goutils.GetNow().Add(item.TTL)
		return item.Payload
	}
}

func (c *MemCache) UpsertWithTTL(key string, val interface{}, ttl time.Duration) {
	i := c.items.Get(key)
	if i != nil {
		item := i.(*cachedItem)
		if item == nil {
			goutils.LogError("Unexpected values in memcache")
			return
		}
		item.Payload = val
		item.TTL = ttl
		item.ExpireTime = goutils.GetNow().Add(ttl)
	} else {
		c.items.Set(key, &cachedItem{
			Payload:    val,
			ExpireTime: goutils.GetNow().Add(ttl),
			TTL:        ttl,
		})
	}
}

func (c *MemCache) Stop() {
	c.ticker.Stop()
	c.items.Unwrap()
}
