package main

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache[V any] struct {
	cache *cache.Cache
}

func (c *Cache[V]) Get(key string) (V, bool) {
	v, ok := c.cache.Get(key)
	if ok {
		return v.(V), true
	}
	var defaultValue V
	return defaultValue, false
}

func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *Cache[V]) Flush() {
	c.cache.Flush()
}

type User struct {
	Name string
}

var userCache = Cache[[]User]{
	cache: cache.New(cache.NoExpiration, cache.NoExpiration),
}

func main() {
	users := []User{{Name: "karamaru"}}
	userCache.Set("users", users, time.Second*1)
	fmt.Println(userCache.Get("users"))
}
