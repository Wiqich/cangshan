package unmarshaler

import (
	"regexp"
	"strings"
)

var (
	tagPattern = regexp.MustCompile("^[^:]+(:\".*\")?$")
)

func LoadStructTag(text string) map[string]string {
	tags := make(map[string]string)
	start := 0
	inString := false
	for i, ch := range text {
		switch ch {
		case ' ':
			if !inString {
				tag := text[start:i]
				if tagPattern.MatchString(tag) {
					if pos := strings.Index(tag, ":"); pos > 0 {
						tags[tag[:pos]] = tag[pos+2 : len(tag)-1]
					} else {
						tags[tag] = ""
					}
				}
			}
			start = i + 1
		case '"':
			inString = !inString
		}
	}
	if start < len(text) {
		tag := text[start:]
		if tagPattern.MatchString(tag) {
			if pos := strings.Index(tag, ":"); pos > 0 {
				tags[tag[:pos]] = tag[pos+2 : len(tag)-1]
			} else {
				tags[tag] = ""
			}
		}
	}
	return tags
}
