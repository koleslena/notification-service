package interfaces

type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Error(args ...any)
	Debugf(msg string, fmtParams ...any)
	Infof(msg string, fmtParams ...any)
	Errorf(msg string, fmtParams ...any)
	Level() LogLevel
	// Need to call with defer after using logger
	Sync() error
}

type LogLevel uint8

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelPanic
	LogLevelFatal
)
