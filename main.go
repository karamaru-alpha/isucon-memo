//package main
//
//import (
//	"fmt"
//	"sync"
//	"time"
//
//	"github.com/patrickmn/go-cache"
//)
//
//type Cache[V any] struct {
//	cache *cache.Cache
//}
//
//func (c *Cache[V]) Get(key string) (V, bool) {
//	v, ok := c.cache.Get(key)
//	if ok {
//		return v.(V), true
//	}
//	var defaultValue V
//	return defaultValue, false
//}
//
//func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
//	c.cache.Set(key, value, ttl)
//}
//
//type User struct {
//	Name string
//}
//
//var userCache = Cache[User]{
//	cache: cache.New(cache.DefaultExpiration, -1),
//}
//
//func main() {
//	karamaru := User{Name: "karamaru"}
//	userCache.Set("karamaru", karamaru, time.Second*1)
//	userCache.Get("karamaru")
//}

package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var flight singleflight.Group

// 5人のユーザーが1ｓ間隔で`slowFunc`をcallする
func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := fmt.Sprintf("user%d", i+1)
			v, _, _ := flight.Do("group1", func() (interface{}, error) {
				value := slowFunc(userID)
				return value, nil
			})
			result := v.(int)
			fmt.Printf("result=%d get by userID:%s.\n", result, userID)
		}()
		time.Sleep(time.Second)
	}
	wg.Wait()
	fmt.Println("finish!")
}

// 10sかかる処理
func slowFunc(userID string) int {
	fmt.Printf("slowFunc called by userID:%s.\n", userID)
	time.Sleep(time.Second * 10)
	return 1
}
