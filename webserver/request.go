package webserver

import (
	"net/http"
)

type Request struct {
	*http.Request
	Attr     map[string]interface{}
	Param    map[string]interface{}
	response *http.ResponseWriter
}
