package cache

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

var (
	zeroTime time.Time
)

func init() {
	application.RegisterModulePrototype("FIFOMemoryCache", new(FIFOMemoryCache))
	gob.Register(new(FIFOCacheItem))
}

type FIFOCacheItem struct {
	Key     string
	Value   interface{}
	Expire  time.Time
	keyNode *list.Element
}

func newFIFOMemoryCacheItem(key string, value interface{}, timeout time.Duration, keyNode *list.Element) *FIFOCacheItem {
	var expire time.Time
	if timeout > 0 {
		expire = time.Now().Add(timeout)
	}
	return &FIFOCacheItem{
		Key:     key,
		Value:   value,
		Expire:  expire,
		keyNode: keyNode,
	}
}

type FIFOMemoryCache struct {
	sync.Mutex
	Capacity     int
	Timeout      time.Duration
	DumpPath     string
	DumpInterval time.Duration
	keys         *list.List
	data         map[string]*FIFOCacheItem
}

func (cache *FIFOMemoryCache) Initialize() error {
	cache.keys = list.New()
	cache.data = make(map[string]*FIFOCacheItem)
	if cache.DumpPath != "" && cache.DumpInterval > 0 {
		if err := cache.load(); err != nil {
			return err
		}
	}
	return nil
}

func (cache *FIFOMemoryCache) Set(key string, value interface{}) error {
	logging.Debug("FIFOMemoryCache.Set: %s -> %v", key, value)
	cache.Lock()
	defer cache.Unlock()
	if _, found := cache.data[key]; !found {
		cache.keys.PushBack(key)
	}
	cache.data[key] = newFIFOMemoryCacheItem(key, value, cache.Timeout, cache.keys.Back())
	for cache.keys.Len() > cache.Capacity {
		front := cache.keys.Front()
		delete(cache.data, front.Value.(string))
		cache.keys.Remove(front)
	}
	return nil
}

func (cache *FIFOMemoryCache) Get(key string) (interface{}, bool, error) {
	logging.Debug("FIFOMemoryCache.Get: %s", key)
	cache.Lock()
	defer cache.Unlock()
	if item, found := cache.data[key]; found {
		if item.Expire == zeroTime || item.Expire.After(time.Now()) {
			return item.Value, true, nil
		}
		delete(cache.data, key)
		cache.keys.Remove(item.keyNode)
	}
	return nil, false, nil
}

func (cache *FIFOMemoryCache) Remove(key string) error {
	delete(cache.data, key)
	return nil
}

func (cache *FIFOMemoryCache) load() error {
	cache.Lock()
	defer cache.Unlock()
	if _, err := os.Stat(cache.DumpPath); err != nil {
		if err == os.ErrNotExist {
			return nil
		} else {
			return err
		}
	}
	var items []*FIFOCacheItem

	if content, err := ioutil.ReadFile(cache.DumpPath); err != nil {
		return err
	} else if err := gob.NewDecoder(bytes.NewBuffer(content)).Decode(&items); err != nil {
		return err
	} else {
		for _, item := range items {
			cache.keys.PushBack(item.Key)
		}
	}
	return nil
}
