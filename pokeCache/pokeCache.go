package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entry map[string]CacheEntry
	mutex sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		entry: make(map[string]CacheEntry),
	}
}

func (c *Cache) Set(key string, value []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entry[key] = CacheEntry{
		createdAt: time.Now(),
		val:       value,
	}
	fmt.Println("======Caching result======")
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.entry[key]
	if !exists {
		return nil, false
	}
	fmt.Println("======Accessing Cache======")
	return entry.val, true
}

func (c *Cache) ReapLoop(expire time.Duration) {

	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			c.mutex.Lock()
			for key, entry := range c.entry {
				if time.Since(entry.createdAt) > expire {
					delete(c.entry, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}
