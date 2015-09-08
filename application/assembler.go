package application

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/yangchenxing/cangshan/structs"
)

var (
	initializedModules = make(map[uintptr]bool)
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
	data interface{}
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

func (asm *assembler) alias(config interface{}) {
	var alias []struct {
		Name  string
		Alias string
	}
	if err := structs.Unmarshal(config, &alias); err != nil {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Unmarshal fail: %s", err.Error()))
		return
	}
	for _, alias := range alias {
		if alias.Name == "" || alias.Alias == "" {
			asm.events <- asm.newEvent(doneEvent, errors.New("Missing \"Name\" or \"Alias\""))
			return
		}
		module := asm.getModule(alias.Name)
		asm.Lock()
		asm.modules[alias.Alias] = module
		for _, waiting := range asm.waitings[alias.Alias] {
			asm.events <- asm.newEvent(receiveEvent, nil)
			waiting.ch <- module
		}
		delete(asm.waitings, alias.Alias)
		asm.Unlock()
	}
	asm.events <- asm.newEvent(doneEvent, nil)
}

func (asm *assembler) setConst(config interface{}) {
	var consts []struct {
		Name  string
		Value interface{}
	}
	if err := structs.Unmarshal(config, &consts); err != nil {
		asm.events <- asm.newEvent(doneEvent, fmt.Errorf("Unmarshal fail: %s", err.Error()))
		return
	}
	asm.Lock()
	defer asm.Unlock()
	for _, c := range consts {
		if c.Name == "" || c.Value == nil {
			asm.events <- asm.newEvent(doneEvent, errors.New("Missing \"Name\" or \"Value\""))
			return
		}
		asm.modules[c.Name] = c.Value
		for _, waiting := range asm.waitings[c.Name] {
			asm.events <- asm.newEvent(receiveEvent, nil)
			waiting.ch <- c.Value
		}
		delete(asm.waitings, c.Name)
	}
	asm.events <- asm.newEvent(doneEvent, nil)
}

func initializeModule(module interface{}) error {
	return nil
}
