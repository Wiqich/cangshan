package logging

import (
	"fmt"
	"github.com/chenxing/cangshan/application"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func init() {
	application.RegisterModuleCreater("LogFormatter",
		func() interface{} {
			return new(Formatter)
		})
}

type event struct {
	level     string
	message   string
	file      string
	line      string
	timestamp time.Time
	funcname  string
	attr      map[string]interface{}
}

func newEvent(skip int, level string, attr map[string]interface{}, format string, params ...interface{}) *event {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "?"
		line = 0
	}
	funcname := "?"
	if callerFunc := runtime.FuncForPC(pc); callerFunc != nil {
		funcname = callerFunc.Name()
	}
	return &event{
		level:     level,
		message:   fmt.Sprintf(format, params...),
		file:      file,
		line:      strconv.Itoa(int(line)),
		timestamp: time.Now(),
		funcname:  funcname,
		attr:      attr,
	}
}

type fieldFormatter interface {
	format(*event) string
}

type basicFieldFormatter func(*event) string

func (formatter basicFieldFormatter) format(e *event) string {
	if formatter == nil {
		return ""
	}
	return formatter(e)
}

type attrFieldFormatter string

func (formatter attrFieldFormatter) format(e *event) string {
	if e.attr == nil {
		return "!MISSING!"
	} else if i := e.attr[string(formatter)]; i == nil {
		return "!MISSING!"
	} else {
		switch v := i.(type) {
		case int:
			fallthrough
		case int8:
			fallthrough
		case int16:
			fallthrough
		case int32:
			fallthrough
		case int64:
			return strconv.FormatInt(int64(v), 10)
		case uint:
			fallthrough
		case uint8:
			fallthrough
		case uint16:
			fallthrough
		case uint32:
			fallthrough
		case uint64:
			return strconv.FormatUint(uint64(v), 10)
		case float32:
			fallthrough
		case float64:
			return strconv.FormatFloat(float64(v), 'f', 6, 64)
		case bool:
			return strconv.FormatBool(v)
		case string:
			return v
		case time.Duration:
			return v.String()
		case time.Time:
			return v.Format("2006-01-02:15:04:05-0700")
		default:
			return fmt.Sprintf("!UNSUPPORTED{%s}!", reflect.TypeOf(v).String())
		}
	}
}

var (
	fieldFormatters = map[string]basicFieldFormatter{
		"level":    func(e *event) string { return e.level },
		"message":  func(e *event) string { return e.message },
		"filepath": func(e *event) string { return e.file },
		"filename": func(e *event) string { return filepath.Base(e.file) },
		"line":     func(e *event) string { return e.line },
		"time":     func(e *event) string { return e.timestamp.Format("2006-01-02:15:04:05-0700") },
		"func":     func(e *event) string { return e.funcname },
		"funcname": func(e *event) string { return e.funcname[strings.LastIndex(e.funcname, ".")+1:] },
	}
	fieldRegexp, _ = regexp.Compile("%[0-9a-zA-Z.]+")
)

type Formatter struct {
	Format  string
	pattern string
	fields  []fieldFormatter
}

func (formatter *Formatter) Initialize() error {
	fields := fieldRegexp.FindAllString(formatter.Format, -1)
	formatter.fields = make([]fieldFormatter, len(fields))
	for i, field := range fields {
		if formatter.fields[i] = fieldFormatters[field[1:]]; formatter.fields[i] == nil {
			formatter.fields[i] = attrFieldFormatter(field[1:])
		}
	}
	formatter.pattern = fieldRegexp.ReplaceAllString(formatter.Format, "%s") + "\n"
	return nil
}

func (formatter *Formatter) format(e *event) string {
	fields := make([]interface{}, len(formatter.fields))
	for i, fieldFormatter := range formatter.fields {
		fields[i] = fieldFormatter.format(e)
	}
	return fmt.Sprintf(formatter.pattern, fields...)
}
