package simplerest

import (
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
	"github.com/yangchenxing/cangshan/webserver"
)

func init() {
	application.RegisterModulePrototype("WebServerSimpleREST", new(SimpleREST))
	application.RegisterBuiltinModule("WebServerSimpleRESTOperationIdentifier",
		new(SimpleRESTOperationIdentifier))
}

type Resource interface {
	Get(query map[string]interface{}, request *webserver.Request) (map[string]interface{}, error)
	Search(query map[string]interface{}, request *webserver.Request) ([]map[string]interface{}, error)
	Create(query map[string]interface{}, before, after []Trigger, request *webserver.Request) (map[string]interface{}, error)
	Update(query map[string]interface{}, before, after []Trigger, request *webserver.Request) (map[string]interface{}, error)
}

type Trigger interface {
	Handle(table string, oldEntity, newEntity map[string]interface{}, request *webserver.Request) error
}

type SimpleREST struct {
	Resource Resource
	Triggers struct {
		BeforeCreate []Trigger
		AfterCreate  []Trigger
		BeforeUpdate []Trigger
		AfterUpdate  []Trigger
	}
}

func (handler SimpleREST) Handle(request *webserver.Request) {
	method, ok := request.Attr["simplerest.method"].(string)
	if !ok {
		logging.Error("Missing attribute: simplerest.method")
		request.WriteAndStop(500, nil, "")
	}
	switch method {
	case "Get":
		entity, err := handler.Resource.Get(request.Param, request)
		if err != nil {
			logging.Error("Get entity fail: %s", err.Error())
			webserver.WriteStandardJSONResult(request, false, "message", err.Error())
		} else if entity == nil {
			webserver.WriteStandardJSONResult(request, true, "entities", []interface{}{})
		} else {
			webserver.WriteStandardJSONResult(request, true, "entities", []interface{}{entity})
		}
	case "Search":
		entities, err := handler.Resource.Search(request.Param, request)
		if err != nil {
			logging.Error("Search entity fail: %s", err.Error())
			webserver.WriteStandardJSONResult(request, false, "message", err.Error())
		} else {
			webserver.WriteStandardJSONResult(request, true, "entities", entities)
		}
	case "Create":
		entity, err := handler.Resource.Create(request.Param,
			handler.Triggers.BeforeCreate, handler.Triggers.AfterCreate, request)
		if err != nil {
			logging.Error("Search entity fail: %s", err.Error())
			webserver.WriteStandardJSONResult(request, false, "message", err.Error())
		} else {
			webserver.WriteStandardJSONResult(request, true, "entities", []interface{}{entity})
		}
	case "Update":
		entity, err := handler.Resource.Create(request.Param,
			handler.Triggers.BeforeUpdate, handler.Triggers.AfterUpdate, request)
		if err != nil {
			logging.Error("Update entity fail: %s", err.Error())
			webserver.WriteStandardJSONResult(request, false, "message", err.Error())
		} else {
			webserver.WriteStandardJSONResult(request, true, "entities", []interface{}{entity})
		}
	default:
		logging.Error("Unknown simplerest.method: %s", method)
		request.WriteAndStop(500, nil, "")
	}
}

type SimpleRESTOperationIdentifier struct{}

func (handler SimpleRESTOperationIdentifier) Handle(request *webserver.Request) {
	switch request.Method {
	case "GET":
		if len(request.Param) > 0 {
			request.Attr["simplerest.method"] = "Get"
			request.Debug("identify simplerest method: Get")
		} else {
			request.Attr["simplerest.method"] = "Search"
			request.Debug("identify simplerest method: Search")
		}
	case "POST":
		if len(request.Param) > 0 {
			request.Attr["simplerest.method"] = "Update"
			request.Debug("identify simplerest method: Update")
		} else {
			request.Attr["simplerest.method"] = "Create"
			request.Debug("identify simplerest method: Create")
		}
	default:
		request.WriteAndStop(405, nil, "")
	}
}
