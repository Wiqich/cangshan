package logging

import (
	"fmt"
	"github.com/chenxing/cangshan/container/poset"
	"github.com/chenxing/cangshan/structs/unmarshaler"
	"io"
	"reflect"
)

type LogWriter io.Writer

type Handler struct {
	Levels    *Levels
	Formatter *Formatter `assemble`
	Writer    LogWriter  `assemble`
}

func (handler Handler) WriteLog(event *Event) error {
	if !handler.Levels.Accept(event.level) {
		return nil
	}
	log := handler.Formatter.Format(event)
	if _, err := handler.LogWriter.Write([]byte(log)); err != nil {
		return fmt.Errorf("write log fail: %s", err.Error())
	}
	return nil
}
