package memcache

import (
	"fmt"
	"time"

	mc "github.com/bradfitz/gomemcache/memcache"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/client/coordination"
	"github.com/yangchenxing/cangshan/client/kv"
	"github.com/yangchenxing/cangshan/logging"
)

const (
	defaultRetry uint = 3
)

func init() {
	application.RegisterModulePrototype("Memcache", new(Memcache))
}

type Memcache struct {
	*mc.Client
	Coordination coordination.Coordination
	ClusterName  string
	Port         uint
	Timeout      time.Duration
	Retry        uint
}

func (client *Memcache) Initialize() error {
	var servers []string
	if nodes, err := client.Coordination.Discover(client.ClusterName); err != nil {
		return err
	} else {
		servers = make([]string, len(nodes))
		for i, node := range nodes {
			servers[i] = fmt.Sprintf("%s:%d", node.Key, client.Port)
		}
	}
	logging.Debug("initalize memcache with servers: %v", servers)
	client.Client = mc.New(servers...)
	if client.Timeout > 0 {
		client.Client.Timeout = client.Timeout
	}
	if client.Retry == 0 {
		client.Retry = defaultRetry
	}
	return nil
}

func (client *Memcache) Ping() error {
	return nil
}

func (client *Memcache) Get(key string) ([]byte, error) {
	var item *mc.Item
	var err error
	for i := uint(0); i < client.Retry; i++ {
		item, err = client.Client.Get(key)
		if err == nil {
			return item.Value, nil
		}
		logging.Warn("Memcache[%s].Get %d/%d fail: key=%v, %s",
			client.ClusterName, i+1, client.Retry, key, err)
	}
	return nil, err
}

func (client *Memcache) GetMulti(keys ...string) (map[string][]byte, error) {
	var items map[string]*mc.Item
	var err error
	for i := uint(0); i < client.Retry; i++ {
		items, err = client.Client.GetMulti(keys)
		if err == nil {
			result := make(map[string][]byte)
			for _, item := range items {
				result[item.Key] = item.Value
			}
			return result, nil
		}
		logging.Warn("Memcache[%s].GetMulti %d/%d fail: keys=%v, %s",
			client.ClusterName, i+1, client.Retry, keys, err)
	}
	return nil, err
}

func (client *Memcache) Set(key string, value []byte, maxage time.Duration) error {
	item := &mc.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(maxage.Seconds()),
	}
	var err error
	for i := uint(0); i < client.Retry; i++ {
		if err = client.Client.Set(item); err == nil {
			return nil
		}
		logging.Warn("Memcache[%s].Set %d/%d fail: key=%s, value=%v, maxage=%v, %s",
			client.ClusterName, i+1, client.Retry, key, value, maxage, err)
	}
	return err
}

func (client *Memcache) SetMulti(items []kv.Item) error {
	var err error
	j := 0
	for i := uint(0); i < client.Retry; i++ {
		for j < len(items) {
			key := items[j].Key
			value := items[j].Value
			maxage := items[j].MaxAge
			if key != "" {
				item := &mc.Item{
					Key:        key,
					Value:      value,
					Expiration: int32(maxage.Seconds()),
				}
				if err = client.Client.Set(item); err != nil {
					logging.Warn("Memcache[%s].SetMulti %v fail: %s",
						client.ClusterName, items, err)
					break
				}
			}
			j++
		}
		if j == len(items) {
			return nil
		}
	}
	return err
}

func (client *Memcache) Remove(key string) error {
	var err error
	for i := uint(0); i < client.Retry; i++ {
		if err = client.Client.Delete(key); err == nil {
			return nil
		}
	}
	return err
}
