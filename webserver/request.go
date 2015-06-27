package webserver

import (
	"bitbucket.org/chenxing/cangshan/logging"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var (
	MultipartMaxMemory = 2048
)

type Request struct {
	*http.Request
	Attr         map[string]interface{}
	Param        map[string]interface{}
	response     http.ResponseWriter
	status       int
	content      bytes.Buffer
	contentType  string
	receiveTime  time.Time
	logFormatter *logging.Formatter
	rewriteCount int
}

func newRequest(request *http.Request, response http.ResponseWriter, formatter *logging.Formatter) *Request {
	req := &Request{
		Request:      request,
		Attr:         make(map[string]interface{}),
		Param:        make(map[string]interface{}),
		response:     response,
		receiveTime:  time.Now(),
		logFormatter: formatter,
	}
	return req
}

func (request *Request) ResponseHeader() http.Header {
	return request.response.Header()
}

func (request *Request) Write(status int, content []byte, contentType string) error {
	request.status = status
	request.content.Reset()
	if content != nil {
		if _, err := request.content.Write(content); err != nil {
			return fmt.Errorf("Write buffer fail: %s", err.Error())
		}
	}
	if contentType != "" {
		request.ResponseHeader().Set("Content-Type", contentType)
	} else {
		request.ResponseHeader().Del("Content-Type")
	}
	return nil
}

func (request *Request) buildResponse() error {
	request.response.WriteHeader(request.status)
	if _, err := request.response.Write(); err != nil {
		return fmt.Errorf("Write response content fail: %s", err.Error())
	}
	return nil
}

func (request *Request) logAccess(access string) {
	request.Attr["request.method"] = request.Method
	request.Attr["request.url"] = request.URL.String()
	request.Attr["request.status"] = request.status
	request.Attr["request.bodylen"] = request.content.Len()
	request.Attr["request.proto"] = request.Proto
	request.Attr["request.timecost"] = time.Now().Sub(request.receiveTime)
	request.Attr["request.time"] = request.receiveTime
	logging.LogEx(2, "access", request.Attr, "")
}

func (request *Request) Debug(format string, params ...interface{}) {
	logging.LogEx(2, "debug", request.logFormatter, request.Attr, format, params...)
}

func (request *Request) Info(format string, params ...interface{}) {
	logging.LogEx(2, "info", request.logFormatter, request.Attr, format, params...)
}

func (request *Request) Warn(format string, params ...interface{}) {
	logging.LogEx(2, "warn", request.logFormatter, request.Attr, format, params...)
}

func (request *Request) Error(format string, params ...interface{}) {
	logging.LogEx(2, "error", request.logFormatter, request.Attr, format, params...)
}

func (request *Request) Fatal(format string, params ...interface{}) {
	logging.LogEx(2, "fatal", request.logFormatter, request.Attr, format, params...)
}
