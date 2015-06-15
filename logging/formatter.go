package logging

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	fieldFormatters = map[string]FieldFormatter{
		"level":            levelFormatter,
		"message":          messageFormatter,
		"filename":         filenameFormatter,
		"line":             lineFormatter,
		"funcname":         funcnameFormatter,
		"funcshortname":    funcshortnameFormatter,
		"time":             timeFormatter("2006-01-02:15:04:05"),
		"time.z":           timeFormatter("2006-01-02:15:04:05-0700"),
		"time.nginx":       timeFormatter("02/Jan/2006:15:04:05 -0700"),
		"time.ansic":       timeFormatter(time.ANSIC),
		"time.unixdate":    timeFormatter(time.UnixDate),
		"time.rubydate":    timeFormatter(time.RubyDate),
		"time.rfc822":      timeFormatter(time.RFC822),
		"time.rfc822z":     timeFormatter(time.RFC822Z),
		"time.rfc850":      timeFormatter(time.RFC850),
		"time.rfc1123":     timeFormatter(time.RFC1123),
		"time.rfc1123z":    timeFormatter(time.RFC1123Z),
		"time.rfc3339":     timeFormatter(time.RFC3339),
		"time.rfc3339nano": timeFormatter(time.RFC3339Nano),
		"time.kitchen":     timeFormatter(time.Kitchen),
		"time.stamp":       timeFormatter(time.Stamp),
		"time.stampmilli":  timeFormatter(time.StampMilli),
		"time.stampmicro":  timeFormatter(time.StampMicro),
		"time.stampnano":   timeFormatter(time.StampNano),
	}

	fieldRegexp, _ = regexp.Compile("\\%\\([^)]+\\)")
)

type FieldFormatter interface {
	Format(*event, string) string
}

type timeFormatter string

func (formatter *timeFormatter) Format(ev *event) string {
	return ev.timestamp.Format(formatter.format)
}

type Formatter struct {
	Format string
	fields []FieldFormatter
}

func (formatter *Formatter) Initialize() error {
	fieldNames := fieldRegexp.FindAllString(formatter.format)
	if fieldNames == nil {
		return nil
	}
	formatter.fields = make([]FieldFormatter, len(fieldNames))
	for i, fieldName := range fieldNames {
		var found bool
		fieldName = fieldName[2 : len(fieldName)-1]
		if formatter.fields[i], found = fieldFormatters[fieldName]; !found {
			if pos := strings.Index(fieldName, "."); pos > 0 {
				if formatter.fields[i] = fieldFormatters[fieldName[:pos+1]]; formatter.fields[i] == nil {
					return fmt.Errorf("unknown field: %s", fieldName)
				}
			}
		}
	}
	formatter.format = fieldRegexp.ReplaceAllString(formatter.format, "%s")
	return nil
}

func (formatter *Formatter) Format(event *Event) string {
	fields := make([]interface{}, len(formatter.fields))
	for i, fieldFormatter := range formatter.fields {
		fields[i] = fieldFormatter(event)
	}
	return fmt.Sprintf(formatter.format, fields...)
}
