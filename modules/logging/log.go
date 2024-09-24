package log

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"
)

const (
	traceStr = "TRACE"
	debugStr = "DEBUG"
	infoStr  = "INFO"
	warnStr  = "WARN"
	offStr   = "OFF"
	panicStr = "PANIC"
	errorStr = "ERROR"
	hackStr  = "HACK"
)

var file *os.File
var mu sync.Mutex

var logLevel uint8 = 4      // by default its trace
var timeEnabled bool = true // this enables execution time measurements

var logConsole bool = false
var logFile bool = false

var logFileReady bool = false

var lastDay int = time.Now().Second()

func SetupLogging(level string, console bool, file bool) {
	logConsole = console
	logFile = file

	makeLogFolder := func() {
		if err := os.MkdirAll("logs", fs.FileMode(os.ModePerm)); err != nil {
			Error(err.Error())
			Fatal("Error creating log folder")
		}
	}

	if logConsole {
		Info("Console logging enabled")
	}

	if logFile {
		makeLogFolder()
		newLogFile()
	}

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
}

func newLogFile() {
	Info("Opening log file for logging...")
	currentTime := time.Now().Format("2006-01-02")
	var err error
	file, err = os.OpenFile(fmt.Sprintf("./logs/%s.log", currentTime), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Error(err.Error())
		Fatal("Error opening log file")
	}
	logFileReady = true
}

func logMsg(logLevelStr string, format string, v ...any) {
	var msg = fmt.Sprintf(format, v...)

	var timestamp = time.Now().Format("2006-01-02, 15:04:05.999")

	var logMsg = fmt.Sprintf("[%s] (%s) %s\n", logLevelStr, timestamp, msg)

	if logConsole || logFile && !logFileReady {
		fmt.Print(logMsg)
	}

	if logFileReady {
		// if time.Now().Minute() != lastDay {
		// 	newLogFile()
		// 	lastDay = time.Now().Minute()
		// }
		file.WriteString(logMsg)
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

func FatalError(err string, format string, v ...any) {
	Error(err)
	logMsg(panicStr, format, v...)
	os.Exit(1)
}
