package xlog

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

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

type Logger struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	Name      string
	lev       Level
	lis       []Listener
	layouters []Layouter
	buf       []byte // for accumulating text to write
}

func New(n string, l Level, lis Listener, layout string) *Logger {
	logger := &Logger{
		Name: n,
		lev:  l,
		lis:  []Listener{lis},
	}
	logger.SetLayout(layout)
	return logger
}

var stdLogger = New("default", InfoLevel, os.Stderr, "%L %D %T %l")

func (l *Logger) SetLayout(layout string) {
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

func (l *Logger) AddListener(lis Listener) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, li := range l.lis {
		if li == lis {
			return false
		}
	}

	if l.lis == nil {
		l.lis = make([]Listener, 0, 1)
	}
	l.lis = append(l.lis, lis)

	return true
}

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

func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Log(PanicLevel, s)
	panic(s)
}
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Log(PanicLevel, s)
	panic(s)
}
func (l *Logger) Fatal(v ...interface{}) {
	l.Log(FatalLevel, v...)
	l.mu.Lock()
	if l.lev >= FatalLevel {
		l.mu.Unlock()
		os.Exit(1)
	}
}
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logf(FatalLevel, format, v...)

	l.mu.Lock()
	if l.lev >= FatalLevel {
		l.mu.Unlock()
		os.Exit(1)
	}
}
func (l *Logger) Error(v ...interface{}) {
	l.Log(ErrorLevel, v...)
}
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logf(ErrorLevel, format, v...)
}
func (l *Logger) Warn(v ...interface{}) {
	l.Log(WarnLevel, v...)
}
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logf(WarnLevel, format, v...)
}
func (l *Logger) Info(v ...interface{}) {
	l.Log(InfoLevel, v...)
}
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(InfoLevel, format, v...)
}
func (l *Logger) Debug(v ...interface{}) {
	l.Log(DebugLevel, v...)
}
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(DebugLevel, format, v...)
}
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

func SetLayout(layout string) {
	stdLogger.SetLayout(layout)
}
func AddListener(lis Listener) bool {
	return stdLogger.AddListener(lis)
}
func RemoveListener(lis Listener) bool {
	return stdLogger.RemoveListener(lis)
}
func Panic(v ...interface{}) {
	stdLogger.Panic(v...)
}
func Panicf(format string, v ...interface{}) {
	stdLogger.Panicf(format, v...)
}
func Fatal(v ...interface{}) {
	stdLogger.Fatal(v...)
}
func Fatalf(format string, v ...interface{}) {
	stdLogger.Fatalf(format, v...)
}
func Error(v ...interface{}) {
	stdLogger.Log(ErrorLevel, v...)
}
func Errorf(format string, v ...interface{}) {
	stdLogger.Logf(ErrorLevel, format, v...)
}
func Warn(v ...interface{}) {
	stdLogger.Log(WarnLevel, v...)
}
func Warnf(format string, v ...interface{}) {
	stdLogger.Logf(WarnLevel, format, v...)
}
func Info(v ...interface{}) {
	stdLogger.Log(InfoLevel, v...)
}
func Infof(format string, v ...interface{}) {
	stdLogger.Logf(InfoLevel, format, v...)
}
func Debug(v ...interface{}) {
	stdLogger.Log(DebugLevel, v...)
}
func Debugf(format string, v ...interface{}) {
	stdLogger.Logf(DebugLevel, format, v...)
}
