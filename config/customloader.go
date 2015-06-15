package config

import (
	"errors"
	"fmt"
	"reflect"
)

type CustomLoader func(in interface{}) (interface{}, error)

var (
	customeLoaders = make(map[string]CustomLoader)
)

type CustomLoaderPlugin struct{}

func (plugin CustomLoaderPlugin) Unmarshal(in, out reflect.Value, tag string) error {
	if loader := customeLoaders[tag]; loader == nil {
		return fmt.Errorf("unknown custom loader: %s", tag)
	} else if instance, err := laoder(in, out); err != nil {
		return fmt.Errorf("load instance fail: %s", err.Error())
	} else {
		out.Set(reflect.ValueOf(instance))
	}
	return nil
}

func RegisterCustomLoader(name string, loader CustomLoader) error {
	if name == "" {
		return errors.New("empty name")
	} else if loader == nil {
		return errors.New("nil loader")
	} else if _, found := customeLoaders[name]; found {
		return fmt.Errorf("duplicated loader: %s", name)
	}
	customeLoaders[name] = loader
	return nil
}
