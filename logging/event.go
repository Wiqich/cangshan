package logging

import (
	"time"
)

type Event struct {
	level         string
	message       string
	filename      string
	line          string
	timestamp     time.Time
	funcname      string
	funcshortname string
	attributes    map[string]string
}

type levelFormatter struct{}

func (formatter levelFormatter) Format(ev *Event, _ string) string {
	return ev.level
}

type messageFormatter struct{}

func (formatter messageFormatter) Format(ev *Event, _ string) string {
	return ev.message
}

type filenameFormatter struct{}

func (formatter filenameFormatter) Format(ev *Event, _ string) string {
	return ev.filename
}

type lineFormatter struct{}

func (formatter lineFormatter) Format(ev *Event, _ string) string {
	return ev.line
}

type funcnameFormatter struct{}

func (formatter funcnameFormatter) Format(ev *Event, _ string) string {
	return ev.funcname
}

type funcshortnameFormatter struct{}

func (formatter funcshortnameFormatter) Format(ev *Event, _ string) string {
	return ev.funcshortname
}
