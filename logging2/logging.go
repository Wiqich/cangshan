package logging

import (
	"container/list"
	"fmt"
	"os"
	"sync"
)

var (
	globalLogging *Logging
	caches        = make(map[string]*list.List)
	flushMutex    sync.Mutex
)

type Logging struct {
	Handlers []*Handler
	handlers map[string][]*Handler
}

func (log *Logging) Initialize() error {
	log.handlers = make(map[string][]*Handler)
	for i, handler := range log.Handlers {
		if err := handler.Initialize(); err != nil {
			return fmt.Errorf("initialize handler[%d] fail: %s", i, err.Error())
		}
		for _, level := range handler.Levels {
			handlers := log.handlers[level]
			if handlers == nil {
				handlers = make([]*Handler, 0, 2)
			}
			log.handlers[level] = append(handlers, handler)
		}
	}
	globalLogging = log
	// fmt.Fprintf(os.Stderr, "initialize logging success: %v\n", log.handlers)
	Flush()
	return nil
}

func Flush() {
	if globalLogging == nil {
		CreateDefaultLogging()
	}
	flushMutex.Lock()
	defer flushMutex.Unlock()
	for level, cache := range caches {
		for elem := cache.Front(); elem != nil; elem = elem.Next() {
			for _, handler := range globalLogging.handlers[level] {
				handler.write(elem.Value.(map[string]string))
			}
		}
		cache.Init()
	}
}

func Log(level string, format string, params ...interface{}) {
	LogSkip(2, level, format, params...)
}

func Debug(format string, params ...interface{}) {
	LogSkip(2, "debug", format, params...)
}

func Info(format string, params ...interface{}) {
	LogSkip(2, "info", format, params...)
}

func Warn(format string, params ...interface{}) {
	LogSkip(2, "warn", format, params...)
}

func Error(format string, params ...interface{}) {
	LogSkip(2, "error", format, params...)
}

func Fatal(format string, params ...interface{}) {
	LogSkip(2, "fatal", format, params...)
}

func LogSkip(skip int, level, format string, params ...interface{}) {
	if globalLogging != nil && len(globalLogging.handlers[level]) == 0 {
		return
	}
	event := newEvent(skip+1, level, format, params...)
	if globalLogging == nil {
		cache := caches[level]
		if cache == nil {
			cache = list.New()
			caches[level] = cache
		}
		cache.PushBack(event)
		// fmt.Fprintf(os.Stderr, "cache log: %s, %v\n", level, event)
	} else {
		for _, handler := range globalLogging.handlers[level] {
			handler.write(event)
		}
		// fmt.Fprintf(os.Stderr, "dispatch log: %s, %v\n", level, event)
	}
}

func CreateDefaultLogging() {
	log := &Logging{
		Handlers: []*Handler{
			&Handler{
				Format: "%level [%time][%file][%line][%func] %message",
				Levels: []string{
					"debug",
					"info",
					"warn",
					"error",
					"fatal",
				},
				Writers: []*writer{
					&writer{
						Type: "stderr",
					},
				},
			},
		},
	}
	if err := log.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "create default logging fail: %s\n", err.Error())
	} else {
		globalLogging = log
	}
}
