package easygate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel string

const (
	DEBUG = LogLevel("DEBUG")
	INFO  = LogLevel("INFO")
	WARN  = LogLevel("WARN")
	ERROR = LogLevel("ERROR")
)

func GetLogLevelFromString(str string) LogLevel {
	switch str {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	}
	return DEBUG
}

type Logger struct {
	impl          zerolog.Logger
	connectWriter *ConnectWriter
}

var instance *Logger = newLog()

type ConnectWriter struct {
	TargetWriter func([]byte) (int, error)
	fileWriter   lumberjack.Logger
}

func newLog() *Logger {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	connectWriter := new(ConnectWriter)
	outputPath := filepath.Join(wd, "easy_gate.log")
	consoleWriter := createLogWriter(connectWriter)
	fileLogger := lumberjack.Logger{
		Filename:   outputPath,
		MaxSize:    200, // MB
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	fileWriter := createLogWriter(&fileLogger)
	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	ins := new(Logger)
	ins.impl = zerolog.New(multi).With().Timestamp().Logger()
	ins.connectWriter = connectWriter
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	return ins
}

func GetLogger() *Logger {
	return instance
}

func (l *Logger) Debug(msg string, v ...interface{}) {
	l.impl.Debug().Msgf(msg, v...)
}

func (l *Logger) Info(msg string, v ...interface{}) {
	l.impl.Info().Msgf(msg, v...)
}

func (l *Logger) Warn(msg string, v ...interface{}) {
	l.impl.Warn().Msgf(msg, v...)
}

func (l *Logger) Error(msg string, v ...interface{}) {
	l.impl.Error().Msgf(msg, v...)
}

func (l *Logger) SetLevel(level LogLevel) {
	switch level {
	case DEBUG:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case INFO:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WARN:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ERROR:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
}

func (l *Logger) SetExternalWriter(w func([]byte) (int, error)) {
	l.connectWriter.TargetWriter = w
}

func (w *ConnectWriter) Write(p []byte) (n int, err error) {
	if w.TargetWriter != nil {
		return w.TargetWriter(p)
	}
	return 0, nil
}

func createLogWriter(w io.Writer) io.Writer {
	return zerolog.ConsoleWriter{
		Out:        w,
		NoColor:    true,
		TimeFormat: "2006-01-02 15:04:05.000",
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
	}

}
