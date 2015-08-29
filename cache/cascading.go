package cache

import (
	"errors"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

var (
	errAllCacheMiss = errors.New("all caches miss")
)

func init() {
	application.RegisterModulePrototype("CascadingCache", new(CascadingCache))
}

type CascadingCache struct {
	Caches []Cache
	depth  int
}

func (cc *CascadingCache) Initialize() error {
	cc.depth = len(cc.Caches) - 1
	return nil
}

func (cc CascadingCache) Set(key string, value interface{}) error {
	logging.Debug("CascadingCache.Set: %s -> %v", key, value)
	err := errAllCacheMiss
	for _, c := range cc.Caches {
		if e := c.Set(key, value); e == nil {
			err = nil
		}
	}
	return err
}

func (cc CascadingCache) Get(key string) (value interface{}, found bool, err error) {
	logging.Debug("CascadingCache.Get: %s", key)
	var hitLevel int
	for i, c := range cc.Caches {
		hitLevel = i
		if value, found, err = c.Get(key); err != nil && i == cc.depth {
			return
		} else if found {
			break
		}
	}
	if found {
		logging.Debug("CascadingCache.Hit: %d", hitLevel)
		for i := hitLevel - 1; i >= 0; i-- {
			cc.Caches[i].Set(key, value)
		}
	}
	return
}

func (cc CascadingCache) Remove(key string) error {
	logging.Debug("CascadingCache.Remove: %s", key)
	for _, c := range cc.Caches {
		c.Remove(key)
	}
	return nil
}
