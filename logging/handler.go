package logging

import (
	"fmt"
	"github.com/chenxing/cangshan/container/poset"
	"github.com/chenxing/cangshan/structs/unmarshaler"
	"io"
)

func init() {
	unmarshaler.RegisterUnmarshalPlugin("cangshan.logging.AcceptedLevels",
		func(config interface{}) *AcceptedLevels {
			whiteCount, blackCount := 0, 0
			levels := &AcceptedLevels{
				levesl: poset.NewStringPoset(),
			}
		})
}

type AcceptedLevels struct {
	blackMode bool
	levels    poset.StringPoset
}

func (levels AcceptedLevels) Accept(level string) bool {

}

type Handler struct {
	Formatter *Formatter      `autoinit:"true"`
	Levels    *AcceptedLevels `customloader:"cangshan.logging.AcceptedLevels"`
	LogWriter io.Writer       `classassemble:"cangshan.logging.LogWriter"`
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
