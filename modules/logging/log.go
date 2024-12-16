package log

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	queryStr      = "QUERY"
	traceStr      = "TRACE"
	debugStr      = "DEBUG"
	infoStr       = "INFO"
	warnStr       = "WARN"
	offStr        = "OFF"
	panicStr      = "PANIC"
	errorStr      = "ERROR"
	hackStr       = "HACK"
	impossibleStr = "IMPOSSIBLE"
)

var logLevel uint8 = 4      // by default its trace
var timeEnabled bool = true // this enables execution time measurements

var logConsole bool = false
var logFile bool = false

var fileLogger *log.Logger

func SetupLogging(level string, console bool, file bool) {
	logConsole = console
	logFile = file

	switch level {
	case traceStr:
		logLevel = 4
	case debugStr:
		logLevel = 3
	case infoStr:
		logLevel = 2
	case warnStr:
		logLevel = 1
	case offStr:
		logLevel = 0
	}

	if logFile {
		logfileRotator := &lumberjack.Logger{
			Filename:   "./logs/log.log",
			MaxSize:    10, // megabytes
			MaxBackups: 100,
			Compress:   true, // disabled by default
		}

		fileLogger = log.New(logfileRotator, "", 0)
		Info("File logging enabled")
	}

	if logConsole {
		Info("Console logging enabled")
	}
}

func logMsg(logLevelStr string, format string, v ...any) {
	var msg = fmt.Sprintf(format, v...)

	var currentTime = time.Now()

	var timestamp = currentTime.Format("2006-01-02 15:04:05")
	var ms = currentTime.Sub(currentTime.Truncate(time.Second)).Milliseconds()

	if logConsole || logFile {
		fmt.Printf("[%s] (%s) (%d ms) %s\n", logLevelStr, timestamp, ms, msg)
	}
	if logFile {
		fileLogger.Printf("%s\t%s\t%d\t%s\n", logLevelStr, timestamp, ms, msg)
	}
}

func Query(format string, v ...any) {
	var replacedQuery = strings.Replace(format, "?", "%v", -1)
	if logLevel >= 4 {
		logMsg(queryStr, replacedQuery, v...)
	}
}

func QueryTx(query string) {
	if logLevel >= 4 {
		logMsg(queryStr, query)
	}
}

func Trace(format string, v ...any) {
	if logLevel >= 4 {
		logMsg(traceStr, format, v...)
	}
}

func Debug(format string, v ...any) {
	if logLevel >= 3 {
		logMsg(debugStr, format, v...)
	}
}

func Info(format string, v ...any) {
	if logLevel >= 2 {
		logMsg(infoStr, format, v...)
	}
}

func Warn(format string, v ...any) {
	if logLevel >= 1 {
		logMsg(warnStr, format, v...)
	}
}

func Time(format string, v ...any) {
	if timeEnabled {
		logMsg("TIME", format, v...)
	}
}

func Hack(format string, v ...any) {
	if logLevel >= 1 {
		logMsg(hackStr, format, v...)
	}
}

func Error(format string, v ...any) {
	logMsg(errorStr, format, v...)
}

func Fatal(format string, v ...any) {
	logMsg(panicStr, format, v...)
	os.Exit(1)
}

func WarnError(err string, format string, v ...any) {
	Error(err)
	logMsg(panicStr, format, v...)
}

func FatalError(err string, format string, v ...any) {
	Error(err)
	logMsg(panicStr, format, v...)
	os.Exit(1)
}

// this is an error logging func that normally should never happen,
// like errors that are never supposed to happen in any way,
// not even accidentally
func Impossible(format string, v ...any) {
	logMsg(impossibleStr, format, v...)
	os.Exit(1)
}
