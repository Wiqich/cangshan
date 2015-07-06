package application

var Log func(level string, format string, params ...interface{})

var Debug func(format string, params ...interface{})

var Info func(format string, params ...interface{})

var Warn func(format string, params ...interface{})

var Error func(format string, params ...interface{})

var Fatal func(format string, params ...interface{})
