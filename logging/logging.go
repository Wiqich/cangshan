package logging

import (
	"container/list"
	"sync"

	"github.com/yangchenxing/cangshan/application"
)

func init() {
	application.RegisterModulePrototype("Logging", new(Logging))
	application.Log = Log
	application.Debug = Debug
	application.Info = Info
	application.Warn = Warn
	application.Error = Error
	application.Fatal = Fatal
}

var (
	globalLogging *Logging
	caches        = make(map[string]*list.List)
	flushMutex    sync.Mutex
	EnableDebug   bool
)

// Logging assemble handlers for logging functions
type Logging struct {
	Handlers []*Handler
	handlers map[string][]*Handler
}

// Initialize the Logging module for applications
func (log *Logging) Initialize() error {
	log.handlers = make(map[string][]*Handler)
	for _, handler := range log.Handlers {
		for _, level := range handler.Levels {
			hs := log.handlers[level]
			if hs == nil {
				hs = make([]*Handler, 0, 1)
			}
			log.handlers[level] = append(hs, handler)
			if level == "debug" {
				EnableDebug = true
			}
		}
	}
	globalLogging = log
	Flush()
	return nil
}

// Log write log with specified level
func Log(level string, format string, params ...interface{}) {
	LogSkip(2, level, format, params...)
}

// Debug write debug log
func Debug(format string, params ...interface{}) {
	LogSkip(2, "debug", format, params...)
}

// Info write info log
func Info(format string, params ...interface{}) {
	LogSkip(2, "info", format, params...)
}

// Warn write warn log
func Warn(format string, params ...interface{}) {
	LogSkip(2, "warn", format, params...)
}

// Error write error log
func Error(format string, params ...interface{}) {
	LogSkip(2, "error", format, params...)
}

// Fatal write fatal log
func Fatal(format string, params ...interface{}) {
	LogSkip(2, "fatal", format, params...)
}

// Flush flush cached log to global Logging instance and clean cache
func Flush() {
	if globalLogging == nil {
		globalLogging = CreateDefaultLogging()
	}
	flushMutex.Lock()
	defer flushMutex.Unlock()
	for level, cache := range caches {
		for e := cache.Front(); e != nil; e = e.Next() {
			for _, handler := range globalLogging.handlers[level] {
				handler.write(e.Value.(event), nil)
			}
		}
		cache.Init()
	}
}

// LogSkip write log with specified level and specified caller by skip argument
func LogSkip(skip int, level string, format string, params ...interface{}) {
	LogEx(skip+1, level, nil, nil, format, params...)
}

// LogEx is the final callee of log writing methods
func LogEx(skip int, level string, formatter *Formatter, attr map[string]interface{}, format string, params ...interface{}) {
	if globalLogging != nil && len(globalLogging.handlers[level]) == 0 {
		return
	}
	e := newEvent(skip+1, level, attr, format, params...)
	if globalLogging == nil {
		cache := caches[level]
		if cache == nil {
			cache = list.New()
			caches[level] = cache
		}
		cache.PushBack(e)
	} else {
		for _, handler := range globalLogging.handlers[level] {
			handler.write(e, formatter)
		}
	}
}

func CreateDefaultLogging() *Logging {
	log := &Logging{
		Handlers: []*Handler{createDefaultHandler()},
	}
	log.Initialize()
	return log
}
