// Package cache provides MemCache as a TTL based simple cache in memory.
package cache

import (
	"errors"
	"time"

	"github.com/hoveychen/go-utils"
	"github.com/hoveychen/go-utils/gomap"
)

type MemCache struct {
	items           *gomap.Map
	ticker          *time.Ticker
	recycleInterval time.Duration
}

type cachedItem struct {
	Payload    interface{}
	ExpireTime time.Time
}

type MemCacheOption func(mc *MemCache)

// NewMemCache returns a memory cache with ttl support.
func NewMemCache(opts ...MemCacheOption) *MemCache {
	c := &MemCache{
		items:           gomap.New(),
		recycleInterval: time.Hour,
	}

	for _, opt := range opts {
		opt(c)
	}

	go func() {
		c.ticker = time.NewTicker(c.recycleInterval)
		for range c.ticker.C {
			c.removeExpired()
		}
	}()
	return c
}

// WithRecycleInterval sets the recycle interval to remove timed out
func WithRecycleInterval(interval time.Duration) MemCacheOption {
	return func(mc *MemCache) {
		mc.recycleInterval = interval
	}
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

// Get returns the active value by given key.
func (c *MemCache) Get(key string) interface{} {
	i, err := c.GetOrError(key)
	if err != nil {
		return nil
	}
	return i
}

// GetOrError returns the active value by given key. If any error occurs,
// like Not Found or Expired, returns err.
// It's the preferred method to check existent, if any value set is nil.
func (c *MemCache) GetOrError(key string) (interface{}, error) {
	i := c.items.Get(key)
	if i == nil {
		return nil, errors.New("Not found")
	} else {
		item := i.(*cachedItem)
		if item == nil {
			return nil, errors.New("Unexpected values")
		}
		if item.ExpireTime.Before(time.Now()) {
			// This item had expired.
			c.items.Delete(key)
			return nil, errors.New("Expired")
		}
		return item.Payload, nil
	}
}

// UpsertWithTTL sets the value by given key into the cache. The k/v expires after
// given ttl duration pass.
func (c *MemCache) UpsertWithTTL(key string, val interface{}, ttl time.Duration) {
	i := c.items.Get(key)
	if i != nil {
		item := i.(*cachedItem)
		if item == nil {
			goutils.LogError("Unexpected values in memcache")
			return
		}
		item.Payload = val
		item.ExpireTime = time.Now().Add(ttl)
	} else {
		c.items.Set(key, &cachedItem{
			Payload:    val,
			ExpireTime: time.Now().Add(ttl),
		})
	}
}

// Stop release potential memory use.
func (c *MemCache) Stop() {
	c.ticker.Stop()
	c.items.Unwrap()
}
