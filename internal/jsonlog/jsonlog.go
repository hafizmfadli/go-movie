package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level type to represent the severity level for a log entry
type Level int8

const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// String return human friendly string for the severity level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger type holds the output destination that the log entries
// will be written to, the minimun severity level that log entries
// will be written for, plus a mutex for coordinating writes.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// NewLogger return a new Logger instance which writes log entries at or above
// a minimum severity level to a specific output destination
func NewLogger(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// PrintInfo encoding message and properties to JSON and then write it to out (io.Writer).
// PrintInfo has LevelInfo severity
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// PrintError encoding err and properties to JSON and then write it to out (io.Writer).
// PrintError has LevelError severity
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// PrintFatal encoding err and properties to JSON and then write it to out (io.Writer).
// PrintFatal has LevelFatal severity
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

// print is internal method for writing the log entry
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	// anonymous struct to holding the data for the log entry
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Include a stack trace for entries at the ERROR and FATAL level
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// variable for holding the actual log entry text
	var line []byte

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	// Lock the mutex so that no two writes to the output destination can happen
	// concurrently. If we don't do this, it's possible that the text for two or more
	// log entries will be intermingled in the output
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))
}

// Implement a Write() method on our logger type so that it satisfies the
// io.Writer interface.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
