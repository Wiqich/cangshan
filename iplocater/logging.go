package iplocater

var Debug func(format string, params ...interface{})

func init() {
	Debug = defaultDebug
}

func defaultDebug(format string, params ...interface{}) {}
