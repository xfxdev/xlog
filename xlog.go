// Package xlog implements a simple logging package.
//
// The Logger provide some methods for formatting log message output and log filter by set log level.
// The Layouter used to format the log message according to the format you want.
//
// It also has a predefined 'standard' Logger accessible through helper functions
// which are easier to use than creating a Logger manually.
//
// The Fatal[f] functions call os.Exit(1) after writing the log message.
// The Panic[f] functions call panic after writing the log message.
package xlog

import (
	"fmt"
	"os"
	"runtime"
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

// Level2Str used to conver Level value to string.
var Level2Str = [...]string{
	"PANIC",
	"FATAL",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
}

// DefaultLoggerLayout give the default log layout.
// "%L %D %T %l" mean "[INFO] 2017/01/05 18:02:17 some log msg..."
// See Layouter for details.
const DefaultLoggerLayout = "%L %D %T %l"

// ParseLevel used to parse string to Level value.
// It will return (ParsedLevel, true) if parse successed, otherwise will return (InfoLevel, false).
func ParseLevel(str string) (Level, bool) {
	for i, v := range Level2Str {
		if v == str {
			return Level(i), true
		}
	}
	return InfoLevel, false
}

// A Logger represents an active logging object that generates lines of
// output to log listeners. A Logger can be used simultaneously from
// multiple goroutines; it guarantees to serialize access to the Writer.
type Logger struct {
	mu             sync.Mutex // ensures atomic writes; protects the following fields
	lev            Level      // log level
	lis            []Listener // log listeners
	layouters      []Layouter // log layouters
	buf            []byte     // for accumulating text to write
	needCallerInfo bool       // flag of caller info need or not
}

// New creates a new Logger.
func New(lev Level, lis Listener, layout string) *Logger {
	logger := &Logger{
		lev: lev,
		lis: []Listener{lis},
	}
	logger.SetLayout(layout)
	return logger
}

var stdLogger = New(InfoLevel, os.Stderr, DefaultLoggerLayout)

// SetLevel set the log level for Logger.
func (l *Logger) SetLevel(lev Level) {
	if lev != l.lev {
		l.mu.Lock()
		l.lev = lev
		l.mu.Unlock()
	}
}

// SetLayout set the layout of log message.
// will use DefaultLoggerLayout by default if layout parameter if empty.
// see Layouter for details.
func (l *Logger) SetLayout(layout string) {
	if len(layout) == 0 {
		layout = DefaultLoggerLayout
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	// clear before set.
	l.layouters = nil

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
				switch layouter.(type) {
				case *layouterFile, *layouterShortFile, *layouterLine:
					l.needCallerInfo = true
				}
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
			if len(layout) > 0 {
				l.layouters = append(l.layouters, &layouterPlaceholder{
					placeholder: layout,
				})
			}
			break
		}
	}
}

// AddListener add a listener to the Logger, return false if the listener existed already, otherwise return true.
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
			// keep listeners's order.
			l.lis = append(l.lis[:i], l.lis[i+1:]...)
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
	l.Log(FatalLevel, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf print a FatalLevel message to the logger followed by a call to os.Exit(1).
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Log(FatalLevel, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Error print a ErrorLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprint(v...))
}

// Errorf print a ErrorLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprintf(format, v...))
}

// Warn print a WarnLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(v ...interface{}) {
	l.Log(WarnLevel, fmt.Sprint(v...))
}

// Warnf print a WarnLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Log(WarnLevel, fmt.Sprintf(format, v...))
}

// Info print a InfoLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	l.Log(InfoLevel, fmt.Sprint(v...))
}

// Infof print a InfoLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Log(InfoLevel, fmt.Sprintf(format, v...))
}

// Debug print a DebugLevel message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(v ...interface{}) {
	l.Log(DebugLevel, fmt.Sprint(v...))
}

// Debugf print a DebugLevel message to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Log(DebugLevel, fmt.Sprintf(format, v...))
}

// Log print a leveled message to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Log(level Level, msg string) {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lev >= level {
		l.buf = l.buf[:0]

		if l.needCallerInfo {
			// release lock while getting caller info - it's expensive.
			l.mu.Unlock()
			var ok bool
			_, file, line, ok = runtime.Caller(2)
			if !ok {
				file = "???"
				line = 0
			}
			// relock
			l.mu.Lock()
		}

		for _, layouter := range l.layouters {
			layouter.layout(&l.buf, level, msg, now, file, line)
		}
		if len(l.buf) == 0 || l.buf[len(l.buf)-1] != '\n' {
			l.buf = append(l.buf, '\n')
		}

		for _, lis := range l.lis {
			lis.Write(l.buf)
		}
	}
}

// SetLevel is equivalent to Logger.SetLevel.
func SetLevel(lev Level) {
	stdLogger.SetLevel(lev)
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
	s := fmt.Sprint(v...)
	stdLogger.Log(PanicLevel, s)
	panic(s)
}

// Panicf is equivalent to Logger.Panicf.
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	stdLogger.Log(PanicLevel, s)
	panic(s)
}

// Fatal is equivalent to Logger.Fatal.
func Fatal(v ...interface{}) {
	stdLogger.Log(FatalLevel, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Logger.Fatalf.
func Fatalf(format string, v ...interface{}) {
	stdLogger.Log(FatalLevel, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Error is equivalent to Logger.Error.
func Error(v ...interface{}) {
	stdLogger.Log(ErrorLevel, fmt.Sprint(v...))
}

// Errorf is equivalent to Logger.Errorf.
func Errorf(format string, v ...interface{}) {
	stdLogger.Log(ErrorLevel, fmt.Sprintf(format, v...))
}

// Warn is equivalent to Logger.Warn.
func Warn(v ...interface{}) {
	stdLogger.Log(WarnLevel, fmt.Sprint(v...))
}

// Warnf is equivalent to Logger.Warnf.
func Warnf(format string, v ...interface{}) {
	stdLogger.Log(WarnLevel, fmt.Sprintf(format, v...))
}

// Info is equivalent to Logger.Info.
func Info(v ...interface{}) {
	stdLogger.Log(InfoLevel, fmt.Sprint(v...))
}

// Infof is equivalent to Logger.Infof.
func Infof(format string, v ...interface{}) {
	stdLogger.Log(InfoLevel, fmt.Sprintf(format, v...))
}

// Debug is equivalent to Logger.Debug.
func Debug(v ...interface{}) {
	stdLogger.Log(DebugLevel, fmt.Sprint(v...))
}

// Debugf is equivalent to Logger.Debugf.
func Debugf(format string, v ...interface{}) {
	stdLogger.Log(DebugLevel, fmt.Sprintf(format, v...))
}

// Log is equivalent to Logger.Log.
func Log(level Level, v ...interface{}) {
	stdLogger.Log(level, fmt.Sprint(v...))
}

// Logf is equivalent to Logger.Logf.
func Logf(level Level, format string, v ...interface{}) {
	stdLogger.Log(level, fmt.Sprintf(format, v...))
}
