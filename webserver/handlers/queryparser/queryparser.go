package queryparser

import (
	"encoding/json"
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
	"io/ioutil"
	"strings"
)

func init() {
	application.RegisterModulePrototype("WebServerQueryParser", new(QueryParserHandler))
}

type QueryParserHandler struct {
	MultipartMaxMemory int64
}

func (handler *QueryParserHandler) Initialize() error {
	if handler.MultipartMaxMemory == 0 {
		handler.MultipartMaxMemory = 10240
	}
	return nil
}

func (handler QueryParserHandler) Handle(request *webserver.Request) {
	var err error
	switch request.Method {
	case "GET":
		fallthrough
	case "HEAD":
		for key, values := range request.URL.Query() {
			if len(values) != 0 {
				request.Param[key] = values[0]
			}
		}
	case "POST":
		fallthrough
	case "PUT":
		contentType := request.Header.Get("Content-Type")
		if contentType != "" {
			contentType = strings.ToLower(strings.Split(contentType, ";")[0])
		}
		switch contentType {
		case "application/json":
			fallthrough
		case "text/json":
			form := make(map[string]interface{})
			if requestBody, err := ioutil.ReadAll(request.Body); err != nil {
				request.Error("Decode JSON query fail: %s", err.Error())
			} else if err = json.Unmarshal(requestBody, &form); err != nil {
				request.Error("Decode JSON query fail: %s", err.Error())
			} else {
				for key, value := range form {
					request.Param[key] = value
				}
				request.Debug("Decode JSON success")
			}
		case "multipart/form-data":
			if err = request.ParseMultipartForm(handler.MultipartMaxMemory); err != nil {
				request.Error("Parse multipart/form-data query fail: %s", err.Error())
			} else {
				for key, values := range request.MultipartForm.Value {
					if len(values) > 0 {
						request.Param[key] = values[0]
					}
				}
				request.Debug("Decode multipart/form-data query success")
			}
		default:
			if err = request.ParseForm(); err != nil {
				request.Error("解析x-www-form-urlencoded表单出错: error=\"%s\"", err.Error())
			} else {
				for key, values := range request.Form {
					if len(values) > 0 {
						request.Param[key] = values[0]
					}
				}
				request.Debug("Decode x-www-form-urlencoded query success")
			}
		}
	default:
		request.Warn("Unsupported http method: %s", request.Method)
	}
}
