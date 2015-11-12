package jaml

import (
	"bytes"
	"container/list"
	"fmt"
	"reflect"
	"time"
)

var (
	types = map[string]reflect.Type{
		"time.Time":     reflect.TypeOf(time.Time{}),
		"time.Duration": reflect.TypeOf(time.Second),
	}
)

func RegisterType(i interface{}) {
	t := reflect.TypeOf(i)
	for t.Kind() == reflect.Interface || t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if pkgPath := t.PkgPath(); pkgPath != "" {
		types[pkgPath+"."+t.Name()] = t
	} else {
		types[t.Name()] = t
	}
}

func create(typeName string) interface{} {
	if t, found := types[typeName]; found {
		return reflect.New(t)
	}
	return nil
}

type namedModule struct {
	name  string
	value interface{}
	err   error
}

type moduleDepInfo struct {
	self string
	dep  string
	ch   chan<- interface{}
}

func Build(v *Value) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	mods := make(chan namedModule)
	deps := make(chan moduleDepInfo)
	waitList := list.New()
	count := 0
	block := 0
	for name, value := range v.Fields {
		if flag, found := flags[name]; found {
			go func() {
				err := setFlag(flag, value)
				mods <- namedModule{name, reflect.ValueOf(flag).Elem().Interface(), err}
			}()
		} else if value.Type != "" {
			go func() {
				v, err := makeModule(name, value, deps)
				mods <- namedModule{name, v, err}
			}()
		}
		count++
	}
	for count > 0 && block < count {
		select {
		case mod := <-mods:
			if mod.err != nil {
				return nil, fmt.Errorf("build module %s fail: %s", mod.name, mod.err.Error())
			}
			res[mod.name] = mod.value
			for elem := waitList.Front(); elem != nil; elem = elem.Next() {
				if elem.Value.(moduleDepInfo).dep == mod.name {
					elem.Value.(moduleDepInfo).ch <- mod.value
					block--
					waitList.Remove(elem)
				}
			}
			count--
		case dep := <-deps:
			if mod, found := res[dep.dep]; found {
				dep.ch <- mod
			} else {
				waitList.PushBack(dep)
				block++
			}
		}
	}
	if count > 0 {
		var buf bytes.Buffer
		for elem := waitList.Front(); elem != nil; elem = elem.Next() {
			if elem != waitList.Front() {
				buf.WriteRune(',')
			}
			buf.WriteString(elem.Value.(moduleDepInfo).self)
			buf.WriteString("->")
			buf.WriteString(elem.Value.(moduleDepInfo).dep)
		}
		return nil, fmt.Errorf("block by dependency: %s", buf.String())
	}
	return res, nil
}

func makeModule(name string, value *Value, deps chan<- moduleDepInfo) (interface{}, error) {
	mod := create(value.Type)
	if mod == nil {
		return nil, fmt.Errorf("unknown type: %q", value.Type)
	}
	if err := value.Unmarshal(reflect.ValueOf(mod), name, deps); err != nil {
		return nil, err
	}
	return mod, nil
}
