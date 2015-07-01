package queryparser

import (
	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/webserver"
	"io/ioutil"
)

func init() {
	application.RegisterModulePrototype("WebServerQueryParser", new(QueryParserHandler))
}

type QueryParserHandler struct {
	MultipartMaxMemory int
}

func (handler *QueryParserHandler) Initialize() error {
	if handler.MultipartMaxMemory == 0 {
		handler.MultipartMaxMemory = 10240
	}
}

func (handler QueryParserHandler) Handle(request *webserver.Request) {
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
			var requestBody []byte
			if requestBody, err = ioutil.ReadAll(request.Body); err != nil {
				request.Error("Decode JSON query fail: %s", err.Error())
				return
			}
			if err = json.Unmarshal(requestBody, &form); err != nil {
				request.Error("Decode JSON query fail: %s", err.Error())
				return
			}
			request.Debug("Decode JSON success")
		case "multipart/form-data":
			if err = request.ParseMultipartForm(handler.MultipartMaxMemory); err != nil {
				request.LogError("Parse multipart/form-data query fail: %s", err.Error())
				return
			}
			for key, values := range request.MultipartForm.Value {
				if len(values) > 0 {
					form[key] = values[0]
				}
			}
			request.Debug("Decode multipart/form-data query success")
		default:
			if err = request.HTTPRequest.ParseForm(); err != nil {
				request.LogError("解析x-www-form-urlencoded表单出错: error=\"%s\"", err.Error())
				return
			}
			for key, values := range request.HTTPRequest.Form {
				if len(values) > 0 {
					form[key] = values[0]
				}
			}
			request.Debug("Decode x-www-form-urlencoded query success")
		}
	default:
		request.Warn("Unsupported http method: %s", request.Method)
	}
}
