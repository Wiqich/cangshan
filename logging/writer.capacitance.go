package logging

import (
	"bytes"
	"container/list"
	"errors"
	"io"
	"sync"
	"time"
)

type CapacitanceWriter struct {
	sync.Once
	Capacity int
	MaxAge   time.Duration
	Receiver io.Writer
	logChan  chan []byte
}

func (writer *CapacitanceWriter) Initialize() error {
	if writer.Capacity == 0 {
		return errors.New("invalid capacity")
	}
	if writer.Receiver == nil {
		return errors.New("no receiver")
	}
	if writer.logChan == nil {
		writer.logChan = make(chan []byte)
	}
	writer.Once.Do(func() {
		go writer.chargeAndEmit()
	})
	return nil
}

func (writer *CapacitanceWriter) Write(p []byte) (int, error) {
	writer.logChan <- p
	return len(p), nil
}

func (writer *CapacitanceWriter) chargeAndEmit() {
	type timedLog struct {
		timestamp time.Time
		log       []byte
	}
	buffer := list.New()
	for {
		log := <-writer.logChan
		if log != nil {
			buffer.PushBack(&timedLog{time.Now(), log})
		}
		if writer.MaxAge > 0 {
			threshold := time.Now().Add(writer.MaxAge * -1)
			for item := buffer.Front(); item != nil; item = item.Next() {
				if item.Value.(*timedLog).timestamp.Before(threshold) {
					buffer.Remove(item)
				}
				break
			}
		}
		if buffer.Len() >= writer.Capacity {
			var content bytes.Buffer
			for item := buffer.Front(); item != nil; item = item.Next() {
				content.Write(item.Value.(*timedLog).log)
			}
			writer.Receiver.Write(content.Bytes())
			buffer.Init()
		}
	}
}
