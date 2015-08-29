package logging

import (
	"container/list"
	"runtime"
	"sync"
	"time"

	"github.com/yangchenxing/cangshan/strings"
)

const (
	defaultCapacitanceTriggerMaxAge = time.Hour * 24
)

type CapacitanceTrigger struct {
	sync.Once
	sync.Mutex
	*list.List
	Capacity uint
	MaxAge   time.Duration
	Debug    bool
}

func (trigger *CapacitanceTrigger) Initialize() error {
	trigger.Once.Do(func() {
		trigger.List = list.New()
		if trigger.MaxAge == 0 {
			trigger.MaxAge = defaultCapacitanceTriggerMaxAge
		}
	})
	return nil
}

func (trigger *CapacitanceTrigger) Trigger() bool {
	trigger.Mutex.Lock()
	defer trigger.Mutex.Unlock()
	if trigger.Capacity <= 1 {
		return true
	}
	now := time.Now()
	var pc uintptr
	var file, funcname string
	var line int
	var ok bool
	if trigger.Debug {
		pc, file, line, ok = runtime.Caller(1)
		if !ok {
			file = "unknown"
			line = 0
		}
		funcname = "unknown"
		if callerFunc := runtime.FuncForPC(pc); callerFunc != nil {
			funcname = callerFunc.Name()
			if match := stringutil.MatchRegexpMap(funcnamePattern, funcname); match != nil {
				if classname, found := match["class"]; found {
					funcname = classname + "." + match["func"]
				} else {
					funcname = match["func"]
				}
			}
		}
		Debug("logging.CapacitanceTrigger.Trigger by %s:%d:%s at %s",
			file, line, funcname, now.Format("2006-01-02:15:04:05-0700"))
	}
	trigger.PushBack(now.Add(trigger.MaxAge))
	for item := trigger.Front(); item != nil; item = trigger.Front() {
		if item.Value.(time.Time).Before(now) {
			if trigger.Debug {
				Debug("logging.CapacitanceTrigger remove expired signal with deadline %s",
					item.Value.(time.Time).Format("2006-01-02:15:04:05-0700"))
			}
			trigger.Remove(item)
		} else {
			break
		}
	}
	if trigger.Len() >= int(trigger.Capacity) {
		trigger.Init()
		return true
	}
	return false
}
