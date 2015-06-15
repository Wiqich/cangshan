package config

import (
	"fmt"
	"github.com/chenxing/cangshan/structs/unmarshaler"
	"reflect"
	"strconv"
)

func init() {
	unmarshaler.RegisterPlugin("autoassemble", new(AutoAssemblerPlugin))
}

type AutoAssemblerPlugin struct{}

func (plugin AutoAssemblerPlugin) Unmarshal(in, out reflect.Value, tag string) error {
	if enabled, err := strconv.ParseBool(tag); err != nil {
		return fmt.Errorf("invalid assembler plugin tag: %s", tag)
	} else if !enabled {
		return unmarshaler.ErrIgnorePlugin
	}
	if in.Kind() != reflect.String {
		return fmt.Errorf("assembler plugin expect in type string, not %s", in.Type())
	} else if out.Kind() != reflect.Ptr {
		return fmt.Errorf("assembler plugin expect out type ptr, not %s", out.Type())
	} else if mod := modules[out.Type().Elem().Name()]; mod == nil {
		return fmt.Errorf("unknown module: %s", out.Type().Elem().Name())
	} else if instance := mod.instances[in.String()]; instance == nil {
		return fmt.Errorf("unknown instance: %s", in.String())
	} else {
		out.Set(reflect.ValueOf(instance))
	}
	return nil
}
