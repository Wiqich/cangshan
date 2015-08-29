package logging

type Logger interface {
	Log(level string, format string, params ...interface{})
	Debug(format string, params ...interface{})
	Info(format string, params ...interface{})
	Warn(format string, params ...interface{})
	Error(format string, params ...interface{})
	Fatal(format string, params ...interface{})
}

type SimpleLogger struct {
	LogHandler   func(level string, format string, params ...interface{})
	DebugHandler func(format string, params ...interface{})
	InfoHandler  func(format string, params ...interface{})
	WarnHandler  func(format string, params ...interface{})
	ErrorHandler func(format string, params ...interface{})
	FatalHandler func(format string, params ...interface{})
}

func (sl SimpleLogger) Log(level string, format string, params ...interface{}) {
	sl.LogHandler(level, format, params...)
}

func (sl SimpleLogger) Debug(format string, params ...interface{}) {
	sl.DebugHandler(format, params...)
}

func (sl SimpleLogger) Info(format string, params ...interface{}) {
	sl.InfoHandler(format, params...)
}

func (sl SimpleLogger) Warn(format string, params ...interface{}) {
	sl.WarnHandler(format, params...)
}

func (sl SimpleLogger) Error(format string, params ...interface{}) {
	sl.ErrorHandler(format, params...)
}

func (sl SimpleLogger) Fatal(format string, params ...interface{}) {
	sl.FatalHandler(format, params...)
}
