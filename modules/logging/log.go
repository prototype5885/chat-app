package log

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
)

const (
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

var mtx sync.Mutex

var logLevel uint8 = 4      // by default its trace
var timeEnabled bool = true // this enables execution time measurements

var logConsole bool = false
var logFile bool = false

var file *os.File
var logFileReady bool = false

var lastTime [3]int = getYearMonthDay()

func SetupLogging(level string, console bool, file bool) {
	logConsole = console
	logFile = file

	if logConsole {
		Info("Console logging enabled")
	}

	if logFile {
		if err := os.MkdirAll("logs", fs.FileMode(os.ModePerm)); err != nil {
			Error(err.Error())
			Fatal("Error creating log folder")
		}
		newLogFile(getYearMonthDay())
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

	logFileReady = false
	compressPreviousLogs()
	logFileReady = true
}

func getYearMonthDay() [3]int {
	year, month, day := time.Now().Date()
	return [3]int{year, int(month), day}
}

func formatFilename(timeStamp [3]int) string {
	return fmt.Sprintf("%d-%d-%d", timeStamp[0], timeStamp[1], timeStamp[2])
}

func newLogFile(timeStamp [3]int) {
	Info("Opening log file for logging...")

	var err error
	file, err = os.OpenFile(fmt.Sprintf("./logs/%s.log", formatFilename(timeStamp)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		FatalError(err.Error(), "Error opening log file")
	}
	Info("Log file opened successfully")
}

func compressPreviousLogs() {
	Debug("Compressing previous logs...")
	// Specify the directory you want to read
	dir := "./logs" // Change to your directory

	var err error

	// Read all files in the directory
	var files []os.DirEntry
	files, err = os.ReadDir(dir)
	if err != nil {
		FatalError(err.Error(), "Error reading log files in log directory")
	}

	var currentLogFileName string = formatFilename(getYearMonthDay()) + ".log"

	var counter int = 0

	// Loop through the files
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") && file.Name() != currentLogFileName {

			var filePath string = fmt.Sprintf("%s/%s", dir, file.Name())

			var logFile *os.File
			logFile, err = os.OpenFile(filePath, os.O_RDONLY, 0666)
			if err != nil {
				FatalError(err.Error(), "Error opening log file [%s] for compression", file.Name())
			}

			data, err := io.ReadAll(bufio.NewReader(logFile))
			if err != nil {
				FatalError(err.Error(), "Error reading log file that needs to be compressed")
			}

			compressedFile, err := os.Create(filePath + ".gz")
			if err != nil {
				FatalError(err.Error(), "Error creating compressed log file")
			}

			writer := gzip.NewWriter(compressedFile)
			_, err = writer.Write(data)
			if err != nil {
				FatalError(err.Error(), "Error writing compressed file")
			}
			err = writer.Close()
			if err != nil {
				FatalError(err.Error(), "Error closing writer that wrote gzip bytes into file")
			}
			err = logFile.Close()
			if err != nil {
				FatalError(err.Error(), "Error closing log file that was compressed")
			}

			// remove previous log file that was compressed
			err = os.Remove(filePath)
			if err != nil {
				FatalError(err.Error(), "Error removing older log file")
			}
			counter++
		}
	}
	Info("Amount of previous log files compressed: [%d]", counter)
}

func logIntoFile(logMsg string) {
	mtx.Lock()
	defer mtx.Unlock()
	var currentTime [3]int = getYearMonthDay()
	if currentTime != lastTime {
		file.Close()
		// logFileReady needs to be disabled here else it will cause a loop
		// when trying to log inside newLogFile()
		logFileReady = false
		compressPreviousLogs()
		Info("New timestamp: [%d], last timestamp: [%d]", currentTime, lastTime)
		newLogFile(currentTime)
		lastTime = currentTime
		// logFileReady can be enabled again
		logFileReady = true
	}
	file.WriteString(logMsg)
}

func logMsg(logLevelStr string, format string, v ...any) {
	var msg = fmt.Sprintf(format, v...)

	var currentTime = time.Now()

	var timestamp = currentTime.Format("2006-01-02 15:04:05")
	var ms = currentTime.Sub(currentTime.Truncate(time.Second)).Milliseconds()

	if logConsole || logFile && !logFileReady {
		fmt.Printf("[%s] (%s) (%d ms) %s\n", logLevelStr, timestamp, ms, msg)
	}

	if logFileReady {
		logIntoFile(fmt.Sprintf("%s\t%s\t%d\t%s\n", logLevelStr, timestamp, ms, msg))
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
