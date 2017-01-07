package xlog

import (
	"time"
)

// Layouter used to format log message.
//   %y : year
//   %M : month
//   %d : day
//   %h : hour
//   %m : min
//   %s : second
//   %l : log msg
//   %L : log level
//   %F : file			eg: /a/b/c/d.go
//	 %f : short file	eg: d.go
//   %i : line
//   %D : %y/%M/%d
//   %T : %h:%m:%s
type Layouter interface {
	layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int)
}

var (
	mapLayouter = map[string]Layouter{
		"%y": &logouterYear{},
		"%M": &logouterMonth{},
		"%d": &logouterDay{},
		"%h": &logouterHour{},
		"%m": &logouterMinute{},
		"%s": &logouterSecond{},
		"%l": &layouterMsg{},
		"%L": &layouterLevel{},
		"%F": &layouterFile{},
		"%f": &layouterShortFile{},
		"%i": &layouterLine{},
		"%D": &layouterDate{},
		"%T": &layouterTime{},
	}
)

type logouterYear struct{}
type logouterMonth struct{}
type logouterDay struct{}
type logouterHour struct{}
type logouterMinute struct{}
type logouterSecond struct{}
type layouterMsg struct{}
type layouterLevel struct{}
type layouterFile struct{}
type layouterShortFile struct{}
type layouterLine struct{}
type layouterDate struct{}
type layouterTime struct{}
type layouterPlaceholder struct {
	placeholder string
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (l *logouterYear) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, t.Year(), 4)
}
func (l *logouterMonth) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, int(t.Month()), 2)
}
func (l *logouterDay) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, t.Day(), 2)
}
func (l *logouterHour) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, t.Hour(), 2)
}
func (l *logouterMinute) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, t.Minute(), 2)
}
func (l *logouterSecond) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, t.Second(), 2)
}
func (l *layouterMsg) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, msg...)
}
func (l *layouterLevel) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, '[')
	*buf = append(*buf, Level2Str[lev]...)
	*buf = append(*buf, ']')
}
func (l *layouterFile) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, file...)
}
func (l *layouterShortFile) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	*buf = append(*buf, short...)
}
func (l *layouterLine) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	itoa(buf, line, -1)
}
func (l *layouterDate) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '/')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '/')
	itoa(buf, day, 2)
}
func (l *layouterTime) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
}

func (l *layouterPlaceholder) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, l.placeholder...)
}
