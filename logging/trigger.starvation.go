package logging

import (
	"errors"
	"sync"
	"time"
)

type StarvationTrigger struct {
	sync.Once
	ColdStartDuration time.Duration
	Interval          time.Duration
	Threshold         int
	EbbStart          time.Duration // 低估标记，低谷期间不触发
	EbbEnd            time.Duration
	Callback          func()
	count             int
}

func (trigger *StarvationTrigger) Initialize() error {
	if trigger.ColdStartDuration == 0 {
		return errors.New("ColdStartDuration cannot be missing or 0")
	}
	if trigger.Interval == 0 {
		return errors.New("Interval cannot be missing or 0")
	}
	if trigger.Callback == nil {
		return errors.New("Callback is missing")
	}
	trigger.Once.Do(func() {
		go func() {
			time.Sleep(trigger.ColdStartDuration)
			for {
				if trigger.count <= trigger.Threshold {
					if trigger.EbbStart == 0 && trigger.EbbEnd == 0 {
						trigger.Callback()
					} else {
						hour, min, sec := time.Now().Clock()
						now := time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec) + time.Second
						if now < trigger.EbbStart || now > trigger.EbbEnd {
							trigger.Callback()
						}
					}
				}
				trigger.count = 0
				time.Sleep(trigger.Interval)
			}
		}()
	})
	return nil
}

func (trigger *StarvationTrigger) Feed() {
	trigger.count++
}
