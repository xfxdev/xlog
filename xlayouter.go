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
//   %n : logger name
//   %l : log msg
//   %L : log level
//   %f : file
//   %i : line
//   %D : %y/%M/%d
//   %T : %h:%m:%s
type Layouter interface {
	layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int)
}

var (
	mapLayouter = map[string]Layouter{
		"%D": &layouterDate{},
		"%T": &layouterTime{},
		"%l": &layouterMsg{},
		"%L": &layouterLevel{},
	}
)

type layouterDate struct {
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
func (l *layouterDate) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '/')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '/')
	itoa(buf, day, 2)
}

type layouterTime struct {
}

func (l *layouterTime) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
}

type layouterMsg struct {
}

func (l *layouterMsg) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, msg...)
}

type layouterPlaceholder struct {
	placeholder string
}

func (l *layouterPlaceholder) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, l.placeholder...)
}

type layouterLevel struct {
}

func (l *layouterLevel) layout(buf *[]byte, lev Level, msg string, t time.Time, file string, line int) {
	*buf = append(*buf, '[')
	*buf = append(*buf, level2Str[lev]...)
	*buf = append(*buf, ']')
}
