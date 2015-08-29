package requestcache

import (
	"bytes"
	"container/list"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterModulePrototype("RequestCacheUpdateHandler", new(RequestCacheUpdateHandler))
	application.RegisterBuiltinModule("RequestCacheFetchHandler", new(RequestCacheFetchHandler))
}

var (
	cache    = list.New()
	capacity = 16
	mutex    sync.Mutex
	lastID   int64 = 0
)

func SetCacheCapacity(value int) {
	if value > 0 {
		capacity = value
	}
}

func add(req *request) {
	mutex.Lock()
	defer mutex.Unlock()
	req.id = time.Now().Unix() * 100000
	if req.id < lastID {
		req.id = lastID + 1
	}
	lastID = req.id
	cache.PushBack(req)
	for cache.Len() > capacity {
		cache.Remove(cache.Front())
	}
}

func get(next int64, count int) []*request {
	mutex.Lock()
	defer mutex.Unlock()
	result := make([]*request, 0, count)
	for item := cache.Front(); item != nil; item = item.Next() {
		req := item.Value.(*request)
		if req.id > next {
			result = append(result, req)
		}
	}
	return result
}

type Request struct {
	ID     int64
	Method string
	URL    string
	Header http.Header
	Body   []byte
}

type closableBuffer *bytes.Buffer

func (buffer closableBuffer) Close() error {
	return nil
}

type RequestCacheUpdateHandler struct {
	CacheRate float32
}

func (handler *RequestCacheUpdateHandler) Initialize() error {
	if handler.CacheSize == 0 {
		handler.CacheSize = defaultCacheSize
	}
}

func (handler RequestCacheUpdateHandler) Handle(request *webserver.Request) {
	if rand.Float32() > handler.CacheRate {
		return
	}
	req := &Request{
		Method: request,
		URL:    request.URL.String(),
		Header: request.Header,
		Body:   nil,
	}
	if request.Body != nil {
		if body, err := ioutil.ReadAll(request.Body); err != nil {
			request.Error("read request body fail: %s", err)
		} else {
			req.Body = body
			request.Body = closableBuffer(bytes.NewBuffer(body))
		}
	}
	go add(req)
}

type RequestCacheFetchHandler struct{}

func (handler RequestCacheFetchHandler) Handle(request *webserver.Request) {
	query := request.URL.Query()
	nextID := 0
	count := capacity
	if temp := query.Get("next"); temp != "" {
		if value, err := strconv.Atoi(temp); err != nil {
			request.Warn("invalid parameter: next")
		} else {
			nextID = value
		}
	}
	if temp := query.Get("count"); temp != "" {
		if value, err := strconv.Atoi(temp); err != nil {
			request.Warn("invalid parameter: count")
		} else {
			count = value
		}
	}
	result := get(nextID, count)
	resultContent, _ := json.Marshal(result)
	request.Write(http.StatusOK, resultContent, "text/json")
}
