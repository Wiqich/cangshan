package experiment

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/yangchenxing/cangshan/structs"
)

var (
	routerTypes map[string]reflect.Type
)

type Router interface {
	Initialize() error
	SelectBranch(features Features) (int, error)
}

func RegisterRouterType(name string, typ reflect.Type) {
	routerTypes[name] = typ
}

func createRouter(data map[string]interface{}) (Router, error) {
	if temp, found := data["type"]; !found {
		return nil, nil
	} else if typeName, ok := temp.(string); !ok {
		return nil, errors.New("Type field is not string")
	} else if typ, found := routerTypes[typeName]; !found {
		return nil, fmt.Errorf("Unknown router type: %s", typeName)
	} else {
		routerValue := reflect.New(typ)
		if err := structs.Unmarshal(data, reflect.Indirect(routerValue)); err != nil {
			return nil, err
		}
		router := routerValue.Interface().(Router)
		return router, router.Initialize()
	}
}
