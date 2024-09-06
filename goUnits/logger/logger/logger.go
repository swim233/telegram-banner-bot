package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	logger   *log.Logger
	logFile  *os.File
	logLevel int
	mutex    sync.Mutex
}

var instance *Logger
var once sync.Once

func getInstance() *Logger {
	once.Do(func() {
		instance = &Logger{}
		instance.initLogger()
	})
	return instance
}

func (l *Logger) initLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating logs directory: %v", err)
	}

	logFilePath := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	terminalWriter := &colorWriter{writer: os.Stdout}
	fileWriter := &colorWriter{writer: file, noColor: true}

	multiWriter := io.MultiWriter(terminalWriter, fileWriter)
	l.logger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)
	l.logFile = file
	l.logLevel = LevelInfo
}

func (l *Logger) log(level int, levelStr string, format string, v ...interface{}) {
	if level < l.logLevel {
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var colorStart string
	var colorEnd = "\033[0m"

	switch levelStr {
	case "[DEBUG]":
		colorStart = "\033[32m"
	case "[INFO]":
		colorStart = "\033[34m"
	case "[WARN]":
		colorStart = "\033[33m"
	case "[ERROR]":
		colorStart = "\033[31m"
	default:
		colorStart = "\033[37m"
	}

	msg := fmt.Sprintf(format, v...)
	logMsg := fmt.Sprintf("%s %s", levelStr, msg)

	l.logger.Output(3, fmt.Sprintf("%s%s%s\n", colorStart, logMsg, colorEnd))
}

func Debug(format string, v ...interface{}) {
	getInstance().log(LevelDebug, "[DEBUG]", format, v...)
}

func Info(format string, v ...interface{}) {
	getInstance().log(LevelInfo, "[INFO]", format, v...)
}

func Warn(format string, v ...interface{}) {
	getInstance().log(LevelWarn, "[WARN]", format, v...)
}

func Error(format string, v ...interface{}) {
	getInstance().log(LevelError, "[ERROR]", format, v...)
}

func SetLogLevel(level int) {
	getInstance().logLevel = level
}

func ParseLogLevel(levelStr string) int {
	levelStr = strings.ToUpper(levelStr) // Convert to uppercase
	switch levelStr {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo // default value
	}
}

func Close() {
	if instance != nil && instance.logFile != nil {
		instance.logFile.Close()
	}
}

type colorWriter struct {
	writer  io.Writer
	noColor bool
}

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	if cw.noColor {
		p = stripColors(p)
	}
	return cw.writer.Write(p)
}

func stripColors(p []byte) []byte {
	var result []byte
	inColorCode := false
	for _, b := range p {
		if b == '\033' {
			inColorCode = true
		} else if b == 'm' && inColorCode {
			inColorCode = false
		} else if !inColorCode {
			result = append(result, b)
		}
	}
	return result
}
