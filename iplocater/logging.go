package iplocater

var Debug func(format string, params ...interface{})
var Info func(format string, params ...interface{})
var Warn func(format string, params ...interface{})
var Error func(format string, params ...interface{})
var Fatal func(format string, params ...interface{})

func init() {
	Debug = defaultDebug
	Info = defaultInfo
	Warn = defaultWarn
	Error = defaultError
	Fatal = defaultFatal
}

func defaultDebug(format string, params ...interface{}) {}
func defaultInfo(format string, params ...interface{})  {}
func defaultWarn(format string, params ...interface{})  {}
func defaultError(format string, params ...interface{}) {}
func defaultFatal(format string, params ...interface{}) {}
