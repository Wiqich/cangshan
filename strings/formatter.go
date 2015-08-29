package stringutil

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	verbPattern    = regexp.MustCompile("%\\([0-9a-zA-Z_.]+(:[0-9a-zA-Z_.:]+)?\\)")
	formatterCache = make(map[string]*MapFormatter)
)

type MapFormatter struct {
	pattern string
	keys    []string
}

func NewMapFormatter(pattern string) *MapFormatter {
	keys := make([]string, 0, 8)
	var newPattern bytes.Buffer
	pos := 0
	for _, loc := range verbPattern.FindAllStringIndex(pattern, -1) {
		verb := strings.SplitN(pattern[loc[0]+2:loc[1]-1], ":", 2)
		keys = append(keys, verb[0])
		newPattern.WriteString(pattern[pos:loc[0]])
		if len(verb) == 2 {
			newPattern.WriteRune('%')
			newPattern.WriteString(verb[1])
		} else {
			newPattern.WriteString("%v")
		}
		pos = loc[1]
	}
	if pos < len(pattern) {
		newPattern.WriteString(pattern[pos:])
	}
	return &MapFormatter{
		pattern: newPattern.String(),
		keys:    keys,
	}
}

func (formatter MapFormatter) String() string {
	return fmt.Sprintf("MapFormatter{pattern:%q, keys:%v}", formatter.pattern, formatter.keys)
}

func (formatter MapFormatter) Format(m map[string]interface{}) string {
	params := make([]interface{}, len(formatter.keys))
	for i, key := range formatter.keys {
		params[i] = m[key]
	}
	return fmt.Sprintf(formatter.pattern, params...)
}

func MapFormat(pattern string, m map[string]interface{}) string {
	formatter := formatterCache[pattern]
	if formatter == nil {
		formatter = NewMapFormatter(pattern)
		formatterCache[pattern] = formatter
	}
	return formatter.Format(m)
}

func MapFormatNoCache(pattern string, m map[string]interface{}) string {
	return NewMapFormatter(pattern).Format(m)
}
