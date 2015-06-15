package config

import (
	"errors"
	"fmt"
	"github.com/chenxing/cangshan/container/poset"
	"github.com/chenxing/cangshan/structs/unmarshaler"
	"reflect"
	"strings"
)

var (
	errNoModuleName    = errors.New("module name cannot be empty")
	errNoModule        = errors.New("module cannot be nil")
	errCycleDependency = errors.New("cycle dependency in modules")
)

var (
	modules     = make(map[string]*module)
	modulePoset = poset.NewStringPoset()
)

type module struct {
	fieldName string
	typ       reflect.Type
	instances map[string]interface{}
}

func RegisterModule(fieldName string, typ reflect.Type) error {
	if fieldName == "" {
		return errNoModuleName
	} else if assembler == nil {
		return errNoModule
	}
	modules[typ.Name()] = newModule(fieldName, typ)
	return nil
}

func newModule(fieldName string, typ reflect.Type) {
	return &module{
		fieldName: fieldName,
		typ:       reflect.Type,
		instances: make(map[string]interface{}),
	}
}

func assemblePredefinedModules(configs map[string]interface{}) error {
	for modulePoset.Len() > 0 {
		moduelName := modulePoset.Pop()
		if moduelName == "" {
			return errCycleDependency
		}
		mod := modules[moduelName]
		if mod == nil {
			return fmt.Errorf("unknown module: %s", moduelName)
		}
		if cfg, found := configs[mod.fieldName]; found {
			if err := mod.assembleInstances(cfg.(map[string]interface{})); err != nil {
				return fmt.Errorf("assemble module %s instances fail: %s", moduelName, err.Error())
			}
			instance := reflect.New(mod.typ)
		}
	}
	return nil
}

func (mod *module) assembleInstances(configs map[string]interface{}) error {
	for key, cfg := range configs {
		if instance, err := assembleInstance(cfg.(map[string]interface{}), mod.typ); err != nil {
			return fmt.Errorf("assemble instance %s fail: %s", key, err.Error())
		} else {
			mod.instances[key] = instance
		}
	}
	return nil
}

func (mod *module) assembleInstance(configs map[string]interface{}) (interface{}, error) {
	instance := reflect.New(mod.typ)
	if err := unmarshaler.Unmarshal(cfg.(map[string]interface{}), instance); err != nil {
		return nil, fmt.Errorf("unmarshal fail: %s", err.Error())
	}
	if initMethod, found := mod.typ.MethodByName("Initialize"); found {
		if err := instance.Elem().MethodByName("Initialize").Call([]reflect.Value{})[0]; !err.IsNil() {
			return nil, fmt.Errorf("initial fail: %s", err.Interface().(error).Error())
		}
	}
	return instance.Interface(), nil
}
