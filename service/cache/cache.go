package cache

import (
	gocache "github.com/patrickmn/go-cache"
	"time"
)

var cache *gocache.Cache

func init() {
	cache = gocache.New(720*time.Hour, 720*time.Hour)
}

func Get(key string) interface{} {
	val, exist := cache.Get(key)
	if !exist {
		return nil
	}
	return val
}

func Set(key string, val interface{}) {
	cache.Set(key, val, 100*time.Hour)
}
