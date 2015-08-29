package logging

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/strings"
)

var (
	gopath          = getGopath()
	funcnamePattern = regexp.MustCompile("^((.+/)*[^./]+\\.)?((?P<class>[^./]+)\\.)?(?P<func>[^./]+)$")
	hostname, _     = os.Hostname()
	localIP         = getLocalIPv4()
)

func init() {
	application.RegisterModulePrototype("LogFormatter", new(Formatter))
}

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

type event map[string]interface{}

type logTimestamp time.Time

func (ts logTimestamp) String() string {
	return (time.Time(ts)).Format("2006-01-02:15:04:05-0700")
}

func (ts logTimestamp) Format(s fmt.State, c rune) {
	s.Write([]byte((time.Time(ts)).Format("2006-01-02:15:04:05-0700")))
}

func newEvent(skip int, level string, attr map[string]interface{}, format string, params ...interface{}) event {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "unknown"
		line = 0
	}
	funcname := "unknown"
	if callerFunc := runtime.FuncForPC(pc); callerFunc != nil {
		funcname = callerFunc.Name()
		if match := stringutil.MatchRegexpMap(funcnamePattern, funcname); match != nil {
			if classname, found := match["class"]; found {
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
	ev := map[string]interface{}{
		"loglv":    level,
		"logmsg":   fmt.Sprintf(format, params...),
		"logfile":  file,
		"logline":  line,
		"logtime":  timestamp.Format("2006-01-02:15:04:05-0700"),
		"logfunc":  funcname,
		"hostname": hostname,
		"localip":  localIP,
	}
	for key, value := range attr {
		ev[key] = value
	}
	return ev
}

// A Formatter converts log event to human readable string
type Formatter struct {
	*stringutil.MapFormatter
	Pattern string
}

// Initialize the Formatter module for applications
func (formatter *Formatter) Initialize() error {
	formatter.MapFormatter = stringutil.NewMapFormatter(formatter.Pattern + "\n")
	return nil
}

func getGopath() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	return file[:len(dir)-len("/github.com/yangchenxing/cangshan/logging")]
}
