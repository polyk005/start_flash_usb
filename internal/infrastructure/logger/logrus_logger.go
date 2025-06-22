package logger

import (
	"os"

	"github.com/polyk005/start_flash/internal/domain"
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(level string, format string) *LogrusLogger {
	logger := logrus.New()

	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.SetOutput(os.Stdout)

	switch format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return &LogrusLogger{
		logger: logger,
	}
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LogrusLogger) WithField(key string, value interface{}) domain.Logger {
	return &LogrusLogger{
		logger: l.logger.WithField(key, value).Logger,
	}
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) domain.Logger {
	return &LogrusLogger{
		logger: l.logger.WithFields(logrus.Fields(fields)).Logger,
	}
} 
