// Package xlog implements a simple logging package. It defines a type, Logger,
// with methods for formatting output. It also has a predefined 'standard'
// Logger accessible through helper functions which are easier to use than creating a Logger manually.
// The Fatal[f] functions call os.Exit(1) after writing the log message.
// The Panic[f] functions call panic after writing the log message.
package xlog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Level used to filter log message by the Logger.
type Level uint8

// logging levels.
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var level2Str = [DebugLevel + 1]string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
}

// A Logger represents an active logging object that generates lines of
// output to log listeners. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	name      string     // logger name
	lev       Level      // log level
	lis       []Listener // log listeners
	layouters []Layouter // log layouters
	buf       []byte     // for accumulating text to write
}

// New creates a new Logger.
func New(name string, lev Level, lis Listener, layout string) *Logger {
	logger := &Logger{
		name: name,
		lev:  lev,
		lis:  []Listener{lis},
	}
	logger.SetLayout(layout)
	return logger
}

var stdLogger = New("default", InfoLevel, os.Stderr, "")

// SetLayout set the layout of log message.
// will use '%L %D %T %l' by default if layout parameter if empty.
// see Layouter for detail.
func (l *Logger) SetLayout(layout string) {
	if len(layout) == 0 {
		layout = "%L %D %T %l"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for {
		i := strings.IndexByte(layout, '%')
		if i != -1 {
			if i != 0 {
				l.layouters = append(l.layouters, &layouterPlaceholder{
					placeholder: layout[:i],
				})
			}

			f := layout[i : i+2]
			layouter := mapLayouter[f]
			if layouter != nil {
				l.layouters = append(l.layouters, layouter)
			} else {
				l.layouters = append(l.layouters, &layouterPlaceholder{
					placeholder: f,
				})
			}

			if i+2 > len(layout) {
				break
			}
			layout = layout[i+2:]
		} else {
			break
		}
	}
}

// AddListener add a listener to the Logger, return true if add success, otherwise return false.
func (l *Logger) AddListener(lis Listener) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, li := range l.lis {
		if li == lis {
			return false
		}
	}

	l.lis = append(l.lis, lis)

	return true
}

// RemoveListener remove a listener from the Logger, return true if remove success, otherwise return false.
func (l *Logger) RemoveListener(lis Listener) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, li := range l.lis {
		if li == lis {
			l.lis[i] = l.lis[len(l.lis)-1]
			l.lis = l.lis[:len(l.lis)-1]
			return true
		}
	}
	return false
}

// Panic print a PanicLevel message to the logger followed by a call to panic().
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Log(PanicLevel, s)
	panic(s)
}

// Panicf print a PanicLevel message to the logger followed by a call to panic().
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Log(PanicLevel, s)
	panic(s)
}

// Fatal print a FatalLevel message to the logger followed by a call to os.Exit(1).
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	l.Log(FatalLevel, v...)
	os.Exit(1)
}

// Fatalf print a FatalLevel message to the logger followed by a call to os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logf(FatalLevel, format, v...)
	os.Exit(1)
}

// Error print a ErrorLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.Log(ErrorLevel, v...)
}

// Errorf print a ErrorLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logf(ErrorLevel, format, v...)
}

// Warn print a WarnLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(v ...interface{}) {
	l.Log(WarnLevel, v...)
}

// Warnf print a WarnLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logf(WarnLevel, format, v...)
}

// Info print a InfoLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.Log(InfoLevel, v...)
}

// Infof print a InfoLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(InfoLevel, format, v...)
}

// Debug print a DebugLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(v ...interface{}) {
	l.Log(DebugLevel, v...)
}

// Debugf print a DebugLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(DebugLevel, format, v...)
}

// Log print a leveled message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Log(level Level, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lev >= level {
		s := fmt.Sprint(v...)
		l.buf = l.buf[:0]
		now := time.Now()

		for _, layouter := range l.layouters {
			layouter.layout(&l.buf, level, s, now, "", 0)
		}
		if len(l.buf) == 0 || l.buf[len(l.buf)-1] != '\n' {
			l.buf = append(l.buf, '\n')
		}

		for _, lis := range l.lis {
			lis.Write(l.buf)
		}
	}
}

// Logf print a leveled message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Logf(level Level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lev >= level {
		s := fmt.Sprintf(format, v...)
		l.buf = l.buf[:0]
		now := time.Now()

		for _, layouter := range l.layouters {
			layouter.layout(&l.buf, level, s, now, "", 0)
		}
		if len(l.buf) == 0 || l.buf[len(l.buf)-1] != '\n' {
			l.buf = append(l.buf, '\n')
		}

		for _, lis := range l.lis {
			lis.Write(l.buf)
		}
	}
}

// SetLayout is equivalent to Logger.SetLayout.
func SetLayout(layout string) {
	stdLogger.SetLayout(layout)
}

// AddListener is equivalent to Logger.AddListener.
func AddListener(lis Listener) bool {
	return stdLogger.AddListener(lis)
}

// RemoveListener is equivalent to Logger.RemoveListener.
func RemoveListener(lis Listener) bool {
	return stdLogger.RemoveListener(lis)
}

// Panic is equivalent to Logger.Panic.
func Panic(v ...interface{}) {
	stdLogger.Panic(v...)
}

// Panicf is equivalent to Logger.Panicf.
func Panicf(format string, v ...interface{}) {
	stdLogger.Panicf(format, v...)
}

// Fatal is equivalent to Logger.Fatal.
func Fatal(v ...interface{}) {
	stdLogger.Fatal(v...)
}

// Fatalf is equivalent to Logger.Fatalf.
func Fatalf(format string, v ...interface{}) {
	stdLogger.Fatalf(format, v...)
}

// Error is equivalent to Logger.Error.
func Error(v ...interface{}) {
	stdLogger.Log(ErrorLevel, v...)
}

// Errorf is equivalent to Logger.Errorf.
func Errorf(format string, v ...interface{}) {
	stdLogger.Logf(ErrorLevel, format, v...)
}

// Warn is equivalent to Logger.Warn.
func Warn(v ...interface{}) {
	stdLogger.Log(WarnLevel, v...)
}

// Warnf is equivalent to Logger.Warnf.
func Warnf(format string, v ...interface{}) {
	stdLogger.Logf(WarnLevel, format, v...)
}

// Info is equivalent to Logger.Info.
func Info(v ...interface{}) {
	stdLogger.Log(InfoLevel, v...)
}

// Infof is equivalent to Logger.Infof.
func Infof(format string, v ...interface{}) {
	stdLogger.Logf(InfoLevel, format, v...)
}

// Debug is equivalent to Logger.Debug.
func Debug(v ...interface{}) {
	stdLogger.Log(DebugLevel, v...)
}

// Debugf is equivalent to Logger.Debugf.
func Debugf(format string, v ...interface{}) {
	stdLogger.Logf(DebugLevel, format, v...)
}

// Log is equivalent to Logger.Log.
func Log(level Level, v ...interface{}) {
	stdLogger.Log(level, v...)
}

// Logf is equivalent to Logger.Logf.
func Logf(level Level, format string, v ...interface{}) {
	stdLogger.Logf(level, format, v...)
}
