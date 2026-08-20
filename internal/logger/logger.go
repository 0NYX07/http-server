package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// Logger provides structured logging capabilities
type Logger struct {
	level           LogLevel
	colored         bool
	timestampFormat string
	mu              sync.Mutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global logger
func Init(level string, colored bool, timestampFormat string) {
	once.Do(func() {
		defaultLogger = &Logger{
			level:           parseLogLevel(level),
			colored:         colored,
			timestampFormat: timestampFormat,
		}
	})
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		Init("info", true, "2006-01-02 15:04:05")
	}
	return defaultLogger
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, threadID string, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(l.timestampFormat)
	message := fmt.Sprintf(format, args...)

	var levelStr, color string
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
		color = ColorGray
	case INFO:
		levelStr = "INFO"
		color = ColorGreen
	case WARN:
		levelStr = "WARN"
		color = ColorYellow
	case ERROR:
		levelStr = "ERROR"
		color = ColorRed
	}

	if l.colored {
		if threadID != "" {
			fmt.Fprintf(os.Stdout, "%s[%s]%s %s[%s]%s %s[%s]%s %s\n",
				ColorGray, timestamp, ColorReset,
				ColorCyan, threadID, ColorReset,
				color, levelStr, ColorReset,
				message)
		} else {
			fmt.Fprintf(os.Stdout, "%s[%s]%s %s[%s]%s %s\n",
				ColorGray, timestamp, ColorReset,
				color, levelStr, ColorReset,
				message)
		}
	} else {
		if threadID != "" {
			fmt.Fprintf(os.Stdout, "[%s] [%s] [%s] %s\n", timestamp, threadID, levelStr, message)
		} else {
			fmt.Fprintf(os.Stdout, "[%s] [%s] %s\n", timestamp, levelStr, message)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, "", format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, "", format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, "", format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, "", format, args...)
}

// DebugThread logs a debug message with thread ID
func (l *Logger) DebugThread(threadID string, format string, args ...interface{}) {
	l.log(DEBUG, threadID, format, args...)
}

// InfoThread logs an info message with thread ID
func (l *Logger) InfoThread(threadID string, format string, args ...interface{}) {
	l.log(INFO, threadID, format, args...)
}

// WarnThread logs a warning message with thread ID
func (l *Logger) WarnThread(threadID string, format string, args ...interface{}) {
	l.log(WARN, threadID, format, args...)
}

// ErrorThread logs an error message with thread ID
func (l *Logger) ErrorThread(threadID string, format string, args ...interface{}) {
	l.log(ERROR, threadID, format, args...)
}

// Global convenience functions
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func DebugThread(threadID string, format string, args ...interface{}) {
	GetLogger().DebugThread(threadID, format, args...)
}

func InfoThread(threadID string, format string, args ...interface{}) {
	GetLogger().InfoThread(threadID, format, args...)
}

func WarnThread(threadID string, format string, args ...interface{}) {
	GetLogger().WarnThread(threadID, format, args...)
}

func ErrorThread(threadID string, format string, args ...interface{}) {
	GetLogger().ErrorThread(threadID, format, args...)
}

