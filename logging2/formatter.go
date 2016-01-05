package logging

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	gopath          = getGopath()
	funcnamePattern = regexp.MustCompile("^((.+/)*[^./]+\\.)?((?P<class>[^./]+)\\.)?(?P<func>[^./]+)$")
	hostname, _     = os.Hostname()
	localIP         = getLocalIPv4()
)

func getLocalIPv4() string {
	if infs, err := net.Interfaces(); err == nil && len(infs) > 0 {
		for _, inf := range infs {
			if inf.Flags&net.FlagLoopback != 0 {
				continue
			}
			if addrs, err := inf.Addrs(); err == nil && len(addrs) > 0 {
				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok {
						if ipv4 := ipnet.IP.To4(); ipv4 != nil {
							return ipv4.String()
						}
					}
				}
			}
		}
	}
	return ""
}

func getGopath() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	return file[:len(dir)-len("/github.com/yangchenxing/cangshan/logging2")]
}

func newEvent(skip int, level string, format string, params ...interface{}) map[string]string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "unknown"
		line = 0
	}
	funcname := "unknown"
	if caller := runtime.FuncForPC(pc); caller != nil {
		funcname = caller.Name()
		if match := matchRegexMap(funcnamePattern, funcname); match != nil {
			if classname, found := match["class"]; found && classname != "" {
				funcname = classname + "." + match["func"]
			} else {
				funcname = match["func"]
			}
		}
	}
	if strings.HasPrefix(file, gopath) {
		file = file[len(gopath)+1:]
	}
	timestamp := time.Now()
	event := map[string]string{
		"level":    level,
		"message":  fmt.Sprintf(format, params...),
		"file":     file,
		"line":     strconv.Itoa(line),
		"time":     timestamp.Format("2006-01-02:15:04:05-0700"),
		"func":     funcname,
		"hostname": hostname,
		"localip":  localIP,
	}
	return event
}

func matchRegexMap(pattern *regexp.Regexp, text string) map[string]string {
	submatches := pattern.FindStringSubmatch(text)
	if submatches == nil {
		return nil
	}
	match := make(map[string]string)
	for i, name := range pattern.SubexpNames() {
		match[name] = submatches[i]
	}
	return match
}

type formatter struct {
	pattern string
	names   []string
}

func newFormatter(pattern string) *formatter {
	var patternBuf bytes.Buffer
	var nameBuf bytes.Buffer
	names := make([]string, 0, 16)
	inHolder := false
	for _, r := range []rune(pattern) {
		if inHolder {
			if unicode.IsLetter(r) {
				nameBuf.WriteRune(r)
				continue
			} else {
				patternBuf.WriteString("%s")
				names = append(names, nameBuf.String())
				nameBuf.Reset()
				inHolder = false
			}
		}
		if r == '%' {
			inHolder = true
		} else {
			patternBuf.WriteRune(r)
		}
	}
	if nameBuf.Len() > 0 {
		patternBuf.WriteString("%s")
		names = append(names, nameBuf.String())
		nameBuf.Reset()
		inHolder = false
	}
	patternBuf.WriteRune('\n')
	result := &formatter{
		pattern: patternBuf.String(),
		names:   names,
	}
	// fmt.Fprintf(os.Stderr, "new formatter: %q, %v\n", pattern, result)
	return result
}

func (formatter *formatter) format(event map[string]string) string {
	args := make([]interface{}, len(formatter.names))
	for i, name := range formatter.names {
		args[i] = event[name]
	}
	return fmt.Sprintf(formatter.pattern, args...)
}
