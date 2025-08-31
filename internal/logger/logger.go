package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	logger := logrus.New()
	
	// Настройка формата логов
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	
	// Уровень логирования
	logger.SetLevel(logrus.InfoLevel)
	
	// Вывод в stdout
	logger.SetOutput(os.Stdout)
	
	return &Logger{logger}
}

func (l *Logger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}
