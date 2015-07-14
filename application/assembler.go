package application

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yangchenxing/cangshan/structs"
)

type namedChan struct {
	name string
	ch   chan interface{}
}

func newNamedChan(name string) namedChan {
	return namedChan{name, make(chan interface{})}
}

type assembleEventType int

const (
	waitEvent assembleEventType = iota
	receiveEvent
	doneEvent
)

type assemblerEvent struct {
	typ  assembleEventType
	name string
	err  error
}

type assembler struct {
	*Application
	name string
}

func (asm *assembler) newEvent(typ assembleEventType, err error) *assemblerEvent {
	return &assemblerEvent{
		typ:  typ,
		name: asm.name,
		err:  err,
	}
}

func (asm *assembler) unmarshal(data interface{}, rv reflect.Value) (interface{}, bool, error) {
	if ref, ok := data.(string); ok {
		if strings.HasPrefix(ref, "!REF:") {
			rv.Set(reflect.ValueOf(asm.getModule(ref[5:])))
			return data, true, nil
		} else if strings.HasPrefix(ref, "!CONST:") {
			return asm.getConst(ref[7:]), false, nil
		}
	}
	return data, false, nil
}

func (asm *assembler) loadModule(data interface{}, module interface{}) {
	if err := structs.UnmarshalMapWithHock(data, module, asm.unmarshal); err != nil {
		asm.events <- asm.newEvent(doneEvent, err)
		return
	}
	if initializable, ok := module.(Initializable); ok {
		if err := initializable.Initialize(); err != nil {
			asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Initialize module %s fail: %s", asm.name, err.Error()))
			return
		}
	}
	asm.Lock()
	defer asm.Unlock()
	asm.modules[asm.name] = module
	for _, waiting := range asm.waitings[asm.name] {
		asm.events <- asm.newEvent(receiveEvent, nil)
		waiting.ch <- module
	}
	delete(asm.waitings, asm.name)
	asm.events <- asm.newEvent(doneEvent, nil)
}

func (asm *assembler) getModuleOrWait(name string) (interface{}, <-chan interface{}) {
	if module, found := builtinModules[name]; found {
		return module, nil
	}
	asm.Lock()
	defer asm.Unlock()
	if module, found := asm.modules[name]; found {
		return module, nil
	}
	waitings, found := asm.waitings[name]
	if !found {
		waitings = make([]namedChan, 0, 1)
	}
	ch := newNamedChan(asm.name)
	asm.waitings[name] = append(waitings, ch)
	return nil, ch.ch
}

func (asm *assembler) getModule(name string) interface{} {
	m, c := asm.getModuleOrWait(name)
	if c != nil {
		asm.events <- asm.newEvent(waitEvent, nil)
		m = <-c
	}
	return m
}

func (asm *assembler) alias(config interface{}) {
	if items, ok := config.([]interface{}); !ok {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Invalid alias config: not []interface{}", asm.name))
	} else {
		for _, item := range items {
			if config, ok := item.(map[string]interface{}); !ok {
				asm.events <- asm.newEvent(doneEvent, fmt.Errorf("invalid alias: item type is not map[string]interface{}"))
				return
			} else if name, ok := config["name"].(string); !ok {
				asm.events <- asm.newEvent(doneEvent, fmt.Errorf("invalid alias: missing name"))
				return
			} else if alias, ok := config["alias"].(string); !ok {
				asm.events <- asm.newEvent(doneEvent, fmt.Errorf("invalid alias: missing alias"))
				return
			} else {
				module := asm.getModule(name)
				asm.Lock()
				asm.modules[alias] = module
				for _, waiting := range asm.waitings[alias] {
					asm.events <- asm.newEvent(receiveEvent, nil)
					waiting.ch <- module
				}
				delete(asm.waitings, alias)
				asm.Unlock()
			}
		}
		asm.events <- asm.newEvent(doneEvent, nil)
	}
}

func (asm *assembler) setConst(config interface{}) {
	if config, ok := config.([]interface{}); !ok {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Invalid const config: not []interface{}", asm.name))
	} else {
		asm.Lock()
		defer asm.Unlock()
		asm.consts = make(map[string]interface{})
		for _, item := range config {
			if config, ok := item.(map[string]interface{}); !ok {
				asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Invalid const item: not map[string]interface{}, but with value %v", item))
				return
			} else {
				key, ok := config["key"].(string)
				if !ok {
					asm.events <- asm.newEvent(doneEvent, fmt.Errorf("invalid const item: missing key"))
					return
				}
				asm.consts[key] = config["value"]
			}
		}
		for _, waiting := range asm.waitings["const"] {
			asm.events <- asm.newEvent(receiveEvent, nil)
			waiting.ch <- nil
		}
		delete(asm.waitings, "const")
	}
}

func (asm *assembler) getConst(name string) interface{} {
	asm.Lock()
	if asm.consts == nil {
		waitings, found := asm.waitings["const"]
		if !found {
			waitings = make([]namedChan, 0, 1)
		}
		ch := newNamedChan(asm.name)
		asm.waitings[name] = append(waitings, ch)
		asm.events <- asm.newEvent(waitEvent, nil)
		asm.Unlock()
		<-ch.ch
	} else {
		asm.Unlock()
	}
	return asm.consts[name]
}
