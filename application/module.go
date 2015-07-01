package application

import (
	"reflect"
)

var (
	moduleCreaters = make(map[string]ModuleCreater)
)

type ModuleCreater interface {
	Create() interface{}
}

type Initializable interface {
	Initialize() error
}

type Runable interface {
	Run() error
}

func RegisterModuleCreater(name string, creater ModuleCreater) {
	moduleCreaters[name] = creater
}

func RegisterModuleCreaterFunc(name string, creater func() interface{}) {
	moduleCreaters[name] = moduleFuncCreater(creater)
}

func RegisterModulePrototype(name string, prototype interface{}) {
	typ := reflect.TypeOf(prototype)
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Interface {
		typ = typ.Elem()
	}
	moduleCreaters[name] = &moduleTypeCreater{typ}
}

type moduleFuncCreater func() interface{}

func (creater moduleFuncCreater) Create() interface{} {
	return ((func() interface{})(creater))()
}

type moduleTypeCreater struct {
	typ reflect.Type
}

func (creater moduleTypeCreater) Create() interface{} {
	return reflect.New(creater.typ).Interface()
}
