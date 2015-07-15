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
			module := reflect.ValueOf(asm.getModule(ref[5:]))
			if !module.Type().AssignableTo(rv.Type()) {
				return data, true, fmt.Errorf("module %s of type %s is not assignable to %s",
					ref[5:], module.Type(), rv.Type())
			}
			rv.Set(module)
			return data, true, nil
		} else if strings.HasPrefix(ref, "!CONST:") {
			return asm.getConst(ref[7:]), false, nil
		}
	}
	return data, false, nil
}

func (asm *assembler) loadModule(data interface{}, module interface{}) {
	if err := structs.UnmarshalWithHock(data, module, asm.unmarshal); err != nil {
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
