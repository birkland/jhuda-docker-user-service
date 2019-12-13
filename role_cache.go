package main

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru"
)

const (
	CacheDefaultSize = 100
	CacheDefaultAge  = 1 * time.Minute
)

// RoleCacheConfig configures a role cache
type RoleCacheConfig struct {
	MaxAge  time.Duration // Maximum age before evicting a cache entry
	MaxSize int           // Maximum number of entries
}

// RoleCache caches a limited number of roles, for a specified amount of time.
type RoleCache struct {
	m      sync.Mutex
	config RoleCacheConfig
	cache  *lru.Cache
}

type cacheEntry struct {
	sync.RWMutex
	roles []Role
	ok    bool
	err   error
}

// NewRoleCache initializes a new role cache
func NewRoleCache(cfg RoleCacheConfig) *RoleCache {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = CacheDefaultSize
	}

	if cfg.MaxAge <= 0 {
		cfg.MaxAge = CacheDefaultAge
	}

	cache, _ := lru.New(cfg.MaxSize)

	return &RoleCache{
		cache:  cache,
		config: cfg,
	}
}

// GetOrAdd adds an entry to the cache via invoking the given generator
// function, if there isn't one already.   If there is already  a cache
// entry, it just gets the old cache value, and the roles function is never
// invoked.
//
// The roles function provides the list of roles to cache, possibly performing
// a computation that blocks for a while.  Future to GetOrAdd for the same id
// will block until a value is available or the function
// returns an error. In the case of an error, the value will not be added to
// the cache, and all pending Get requests will return the error
func (c *RoleCache) GetOrAdd(id string, roles func() ([]Role, error)) ([]Role, error) {

	// Fast path: see if we have a cached entry already
	cached, found, err := c.get(id)
	if found {
		return cached, err
	}

	// Critical section.  Global lock, double check that we don't have a cached entry, and create/add a locked one if not
	cached, entry, found, err := func() ([]Role, *cacheEntry, bool, error) {
		c.m.Lock()
		defer c.m.Unlock()

		cached, ok, err := c.get(id)
		if ok {
			return cached, nil, ok, err
		}
		entry := &cacheEntry{}
		entry.Lock()
		c.cache.Add(id, entry)

		return nil, entry, false, nil
	}()

	if found {
		return cached, err
	}

	// OK, now execute the roles function and unlock the cache entry when done.
	defer entry.Unlock()

	if entry.roles, err = roles(); err != nil {
		entry.ok = false
		c.cache.Remove(id)
		return nil, err
	}

	entry.ok = true
	time.AfterFunc(c.config.MaxAge, func() {
		c.cache.Remove(id)
	})

	return entry.roles, nil
}

func (c *RoleCache) get(id string) (roles []Role, ok bool, err error) {
	v, ok := c.cache.Get(id)

	if e, ok := v.(*cacheEntry); ok {
		e.RLock()
		defer e.RUnlock()
		return e.roles, e.ok, e.err
	}

	return nil, ok, nil
}
