package felix

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Logger is the standardized interface for all Loggers used in this project.
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
}

func NewLogger() *DefaultLogger {
	return &DefaultLogger{
		out: os.Stdout,
	}
}

// DefaultLogger is the default logger implementation and logs to stdout.
type DefaultLogger struct {
	out io.Writer
	mu  sync.Mutex
}

// SetOutput sets the output destination for the logger.
func (l *DefaultLogger) SetOutput(w io.Writer) {
	l.out = w
}

// Debug logs with debug level.
func (l *DefaultLogger) Debug(msg string, keyvals ...interface{}) {
	l.log("debug", msg, keyvals...)
}

// Info logs with info level.
func (l *DefaultLogger) Info(msg string, keyvals ...interface{}) {
	l.log("info", msg, keyvals...)
}

// Warn logs with warning level.
func (l *DefaultLogger) Warn(msg string, keyvals ...interface{}) {
	l.log("warn", msg, keyvals...)
}

// Error logs with error level.
func (l *DefaultLogger) Error(msg string, keyvals ...interface{}) {
	l.log("error", msg, keyvals...)
}

func (l *DefaultLogger) log(level, msg string, keyvals ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(keyvals)%2 != 0 {
		// pad keyvals to even number and notify user
		keyvals = append(keyvals, nil, "logerror", "Uneven number of keyval arguments. Please fix!")
	}

	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%-5s [%v] %s", strings.ToUpper(level), time.Now().Format(time.RFC3339), msg)

	for i := 0; i < len(keyvals); i += 2 {
		fmt.Fprintf(b, " %s=%v", keyvals[i], keyvals[i+1])
	}

	fmt.Fprintf(b, "\n")
	l.out.Write(b.Bytes())
}

// NopLogger is a no-op implementation of the Logger interface.
type NopLogger struct{}

// Debug is a no-op implementation of Logger.Debug.
func (NopLogger) Debug(msg string, keyvals ...interface{}) {}

// Info is a no-op implementation of Logger.Info.
func (NopLogger) Info(msg string, keyvals ...interface{}) {}

// Warn is a no-op implementation of Logger.Warn.
func (NopLogger) Warn(msg string, keyvals ...interface{}) {}

// Error is a no-op implementation of Logger.Error.
func (NopLogger) Error(msg string, keyvals ...interface{}) {}
