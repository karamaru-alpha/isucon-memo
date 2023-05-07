package main

import (
	"fmt"
	"sync"
	"time"
)

const CACHE_KEY = "CACHE_KEY"

type Cacher[T any] struct {
	Mutex sync.RWMutex
	Cache map[string]struct {
		Value   T
		Expired time.Time
	}
}

func (c *Cacher[T]) Get(key string) (T, bool) {
	c.Mutex.RLock()
	cache, ok := c.Cache[key]
	c.Mutex.RUnlock()
	if ok && (cache.Expired.IsZero() || time.Now().Before(cache.Expired)) {
		return cache.Value, true
	}
	var defaultValue T
	return defaultValue, false
}

func (c *Cacher[T]) GetAll() []T {
	c.Mutex.RLock()
	slice := make([]T, 0, len(c.Cache))
	for _, v := range c.Cache {
		slice = append(slice, v.Value)
	}
	c.Mutex.RUnlock()
	return slice
}

func (c *Cacher[T]) Set(key string, value T, ttl time.Duration) {
	c.Mutex.Lock()
	var expired time.Time
	if ttl > 0 {
		expired = time.Now().Add(ttl)
	}
	c.Cache[key] = struct {
		Value   T
		Expired time.Time
	}{
		Value:   value,
		Expired: expired,
	}
	c.Mutex.Unlock()
}

func (c *Cacher[T]) Delete(key string) {
	c.Mutex.Lock()
	delete(c.Cache, key)
	c.Mutex.Unlock()
}

func (c *Cacher[T]) Flush() {
	c.Mutex.Lock()
	c.Cache = make(map[string]struct {
		Value   T
		Expired time.Time
	})
	c.Mutex.Unlock()
}

func InitCacher[T any]() Cacher[*T] {
	return Cacher[*T]{
		Cache: make(map[string]struct {
			Value   *T
			Expired time.Time
		}, 0),
	}
}

type Thing struct {
	Name string
}

var ThingCacher = InitCacher[Thing]()

func main() {
	isu := &Thing{Name: "isu"}
	ThingCacher.Set("isu", isu, -1)
	fmt.Println(ThingCacher.Get("isu"))
}

// ↓メソッド拡張
//
//type UserCacher struct {
//	*Cacher[*User]
//}
//
//func (c *UserCacher) IncrementAge(key string) {
//	c.Mutex.Lock()
//	cache, ok := c.Cacher.Cache[key]
//	if !ok {
//		c.Mutex.Unlock()
//		return
//	}
//	cache.Value.Age++
//	c.Mutex.Unlock()
//}
//
//type User struct {
//	Name string
//	Age  int32
//}
//
//var userCache = UserCacher{
//	Cacher: &Cacher[*User]{
//		Cache: make(map[string]struct {
//			Value   *User
//			Expired time.Time
//		}, 0),
//	},
//}
//
//func main() {
//	user := &User{Name: "karamaru"}
//	userCache.Set("karamaru", user, -1)
//	fmt.Println(userCache.Get("karamaru"))
//	userCache.IncrementAge("karamaru")
//	fmt.Println(userCache.Get("karamaru"))
//	time.Sleep(time.Second * 3)
//	fmt.Println(userCache.Get("karamaru"))
//}
