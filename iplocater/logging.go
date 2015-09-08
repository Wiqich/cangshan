package iplocater

var Debug func(format string, params ...interface{})
var Info func(format string, params ...interface{})
var Warn func(format string, params ...interface{})
var Error func(format string, params ...interface{})
var Fatal func(format string, params ...interface{})

func init() {
	if Debug == nil {
		Debug = defaultDebug
	}
	if Info == nil {
		Info = defaultInfo
	}
	if Warn == nil {
		Warn = defaultWarn
	}
	if Error == nil {
		Error = defaultError
	}
	if Fatal == nil {
		Fatal = defaultFatal
	}
}

func defaultDebug(format string, params ...interface{}) {}
func defaultInfo(format string, params ...interface{})  {}
func defaultWarn(format string, params ...interface{})  {}
func defaultError(format string, params ...interface{}) {}
func defaultFatal(format string, params ...interface{}) {}
