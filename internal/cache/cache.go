package cache

import (
	"container/list"
	"sync"
	"time"

	"github.com/lupguo/ip_info/internal/model"
)

type entry struct {
	key     string
	value   *model.IPInfo
	expires time.Time
	elem    *list.Element
}

// Cache is a thread-safe in-memory TTL + LRU cache.
type Cache struct {
	mu      sync.Mutex
	ttl     time.Duration
	maxSize int
	items   map[string]*entry
	lru     *list.List // front = most recently used
}

// New creates a new Cache with the given TTL and maximum size.
func New(ttl time.Duration, maxSize int) *Cache {
	return &Cache{
		ttl:     ttl,
		maxSize: maxSize,
		items:   make(map[string]*entry),
		lru:     list.New(),
	}
}

// Get returns the cached IPInfo for key, or (nil, false) on miss/expiry.
func (c *Cache) Get(key string) (*model.IPInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		c.remove(e)
		return nil, false
	}
	c.lru.MoveToFront(e.elem)
	return e.value, true
}

// Set stores value under key, evicting the LRU entry if at capacity.
func (c *Cache) Set(key string, value *model.IPInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		c.lru.MoveToFront(e.elem)
		e.value = value
		e.expires = time.Now().Add(c.ttl)
		return
	}

	if len(c.items) >= c.maxSize {
		// Evict least recently used
		oldest := c.lru.Back()
		if oldest != nil {
			c.remove(oldest.Value.(*entry))
		}
	}

	e := &entry{
		key:     key,
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	e.elem = c.lru.PushFront(e)
	c.items[key] = e
}

// remove deletes an entry (must be called with lock held).
func (c *Cache) remove(e *entry) {
	delete(c.items, e.key)
	c.lru.Remove(e.elem)
}
