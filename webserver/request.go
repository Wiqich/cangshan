package webserver

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/yangchenxing/cangshan/logging"
)

var (
	RemoteAddrHeaders = []string{"RemoteAddr"}
)

// A Request present a webserver request
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
	done         bool
	stopped      bool
	clientIP     net.IP
}

func newRequest(request *http.Request, response http.ResponseWriter, formatter *logging.Formatter) *Request {
	timestamp := time.Now()
	remoteAddr, _, _ := net.SplitHostPort(request.RemoteAddr)
	if remoteAddr == "" {
		remoteAddr = request.RemoteAddr
	}
	for _, name := range RemoteAddrHeaders {
		if value := request.Header.Get(name); value != "" {
			if host, _, _ := net.SplitHostPort(value); host != "" {
				remoteAddr = host
			} else {
				remoteAddr = value
			}
			break
		}
	}
	clientIP := net.ParseIP(remoteAddr)
	attr := map[string]interface{}{
		"request.remote_addr": clientIP,
		"request.time":        timestamp.Format("[02/Jan/2006:15:04:05 -0700]"),
		"request.method":      request.Method,
		"request.url":         request.URL.String(),
		"request.proto":       request.Proto,
		"request.referer":     request.Referer(),
		"request.user_agent":  request.UserAgent(),
		"request.user":        "-",
		"request.auth":        "-",
		"request.clientip":    clientIP,
	}
	req := &Request{
		Request:      request,
		Attr:         attr,
		Param:        make(map[string]interface{}),
		response:     response,
		receiveTime:  timestamp,
		logFormatter: formatter,
		clientIP:     clientIP,
	}
	return req
}

// ResponseHeader returns response header map that will be sent
func (request *Request) ResponseHeader() http.Header {
	return request.response.Header()
}

// SetCookie add a Set-Cookie header to http response
func (request *Request) SetCookie(cookie *http.Cookie) {
	http.SetCookie(request.response, cookie)
}

func (request *Request) GetClientIP() net.IP {
	return request.clientIP
}

// Write set or overwrite response status, content and content type that will be sent.
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

// Stop the processing of the request and set the response status, content and content type.
// No handler will handle the request after stopped.
func (request *Request) WriteAndStop(status int, content []byte, contentType string) error {
	request.stopped = true
	return request.Write(status, content, contentType)
}

func (request *Request) Stop() {
	request.stopped = true
}

func (request *Request) buildResponse() error {
	if !request.done {
		request.response.WriteHeader(request.status)
		if _, err := request.response.Write(request.content.Bytes()); err != nil {
			return fmt.Errorf("Write response content fail: %s", err.Error())
		}
		request.logAccess()
		request.done = true
	}
	return nil
}

func (request *Request) logAccess() {
	request.Attr["request.timecost"] = time.Now().Sub(request.receiveTime)
	request.Attr["request.status"] = request.status
	request.Attr["request.bodylen"] = request.content.Len()
	logging.LogEx(2, "access", nil, request.Attr, "")
}

// Debug write debug log with web server specified log formatter
func (request *Request) Debug(format string, params ...interface{}) {
	logging.LogEx(2, "debug", request.logFormatter, request.Attr, format, params...)
}

// Info write info log with web server specified log formatter
func (request *Request) Info(format string, params ...interface{}) {
	logging.LogEx(2, "info", request.logFormatter, request.Attr, format, params...)
}

// Warn write warn log with web server specified log formatter
func (request *Request) Warn(format string, params ...interface{}) {
	logging.LogEx(2, "warn", request.logFormatter, request.Attr, format, params...)
}

// Error write error log with web server specified log formatter
func (request *Request) Error(format string, params ...interface{}) {
	logging.LogEx(2, "error", request.logFormatter, request.Attr, format, params...)
}

// Fatal write fatal log with web server specified log formatter
func (request *Request) Fatal(format string, params ...interface{}) {
	logging.LogEx(2, "fatal", request.logFormatter, request.Attr, format, params...)
}

func (request *Request) GetLogger() logging.Logger {
	return &logging.SimpleLogger{
		LogHandler: func(level string, format string, params ...interface{}) {
			logging.LogEx(4, level, request.logFormatter, request.Attr, format, params...)
		},
		DebugHandler: func(format string, params ...interface{}) {
			logging.LogEx(4, "debug", request.logFormatter, request.Attr, format, params...)
		},
		InfoHandler: func(format string, params ...interface{}) {
			logging.LogEx(4, "info", request.logFormatter, request.Attr, format, params...)
		},
		WarnHandler: func(format string, params ...interface{}) {
			logging.LogEx(4, "warn", request.logFormatter, request.Attr, format, params...)
		},
		ErrorHandler: func(format string, params ...interface{}) {
			logging.LogEx(4, "error", request.logFormatter, request.Attr, format, params...)
		},
		FatalHandler: func(format string, params ...interface{}) {
			logging.LogEx(4, "fatal", request.logFormatter, request.Attr, format, params...)
		},
	}
}
