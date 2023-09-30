package services

import (
	"notification-service/inits"
	"notification-service/services/interfaces"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const expectedLoggingCoreSize = 2

func NewLoggerZap(cfg *inits.LogConfig) interfaces.Logger {
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cors := make([]zapcore.Core, 0, expectedLoggingCoreSize)
	level, err := zapcore.ParseLevel(cfg.Level)

	if err != nil {
		panic(err)
	}

	if cfg.IsFileEncoderEnabled {
		path := filepath.Dir(cfg.LogFilePath)
		exist, err := pathExists(path)
		if !exist || err != nil {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				panic("no created logger file directory")
			}
		}
		logFile, _ := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writer := zapcore.AddSync(logFile)
		fileEncoder := zapcore.NewJSONEncoder(zapConfig)
		cors = append(cors, zapcore.NewCore(fileEncoder, writer, level))
	}

	if cfg.IsConsoleEncoderEnabled {
		consoleEncoder := zapcore.NewConsoleEncoder(zapConfig)
		cors = append(cors, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	if len(cors) < 1 {
		panic("No cores for creating logger")
	}

	core := zapcore.NewTee(cors...)
	var logger loggerType

	logger.SugaredLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()

	return logger
}

type loggerType struct {
	*zap.SugaredLogger
}

func (l loggerType) Level() interfaces.LogLevel {
	if l.Desugar().Core().Enabled(zap.PanicLevel) {
		return interfaces.LogLevelPanic
	} else if l.Desugar().Core().Enabled(zap.ErrorLevel) {
		return interfaces.LogLevelError
	}
	return interfaces.LogLevelInfo
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
