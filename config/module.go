package config

import (
	"errors"
	"fmt"
	"github.com/chenxing/cangshan/container/poset"
	"github.com/chenxing/cangshan/structs"
	"reflect"
	"strings"
)

var (
	modules     = make(map[reflect.Type]*Module)
	modulePoset = poset.NewPoset()
)

type Module struct {
	name      string
	typ       reflect.Type
	instances map[string]reflect.Value
}

func NewModule(name string, typ reflect.Type, depends ...*Module) *Module {
	module := &Module{
		name:      name,
		typ:       typ,
		instances: make(map[string]interface{}),
	}
	modules[typ] = module
	for _, dependModule := range depends {
		modulePoset.Add(dependModule, module)
	}
	return module
}

func (module *Module) String() string {
	return module.typ.Name()
}

func assembleModules(conf map[string]interface{}) error {
	for modulePoset.Len() > 0 {
		module := modulePoset.Pop().(*Module)
		var moduleConfig map[string]interface{}
		if temp := conf[module.name]; temp == nil {
			continue
		} else {
			moduleConfig = temp.(map[string]interface{})
		}
		for key, instanceConfig := range moduleConfig {
			instance := reflect.New(module.typ)
			if err := structs.UnmarshalValue(reflect.ValueOf(instanceConfig.(map[string]interface{})), instance); err != nil {
				return fmt.Errorf("assemble modules fail: %s", err.Error())
			}
		}
		structs.RegisterConverter(reflect.TypeOf(""), module.typ, convertStringToModule)
	}
}

func convertStringToModule(in, out reflect.Value) error {
	if module := modules[out.Type()]; module == nil {
		return fmt.Errorf("unknown module: %s", out.Type().Name())
	} else if instance := module.instances[in.String()]; instance == nil {
		return fmt.Errorf("unknown instance: %s@%s", in.String(), out.Type().Name())
	} else {
		out.Set(instance)
		return nil
	}
}
