package unmarshaler

import (
	"errors"
	"fmt"
	"reflect"
)

type UnmarshalPlugin interface {
	Unmarshal(in, out reflect.Value, tag string) error
}

type PostUnmarshalPlugin interface {
	PosUnmarshal(in, out reflect.Value, tag string) error
}

var (
	ErrIgnorePlugin = errors.New("plugin ignored")
)

var (
	unmarshalPlugins     = make(map[string]UnmarshalPlugin)
	postUnmarshalPlugins = make(map[string]PostUnmarshalPlugin)
)

func RegisterUnmarshalPlugin(name string, plugin UnmarshalPlugin) error {
	if name == "" {
		return errors.New("empty name")
	} else if plugin == nil {
		return errors.New("nil plugin")
	} else if _, found := unmarshalPlugins[name]; found {
		return fmt.Errorf("duplicated unmarshal plugin: %s", name)
	}
	unmarshalPlugins[name] = plugin
	return nil
}

func RegisterPostUnmarshalPlugin(name string, plugin PostUnmarshalPlugin) error {
	if name == "" {
		return errors.New("empty name")
	} else if plugin == nil {
		return errors.New("nil plugin")
	} else if _, found := postUnmarshalPlugins[name]; found {
		return fmt.Errorf("duplicated unmarshal plugin: %s", name)
	}
	postUnmarshalPlugins[name] = plugin
	return nil
}
