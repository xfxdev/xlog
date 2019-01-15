package xlog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

type strLogListener struct {
	log string
}

func (lis *strLogListener) Write(p []byte) (n int, err error) {
	lis.log = string(p)
	return n, nil
}

func (l *strLogListener) Close() error {
	return nil
}

func compareInt(t *testing.T, layout Layouter, time time.Time, rv int) {
	var buf []byte
	layout.layout(&buf, DebugLevel, "", time, "", 0)

	lv, err := strconv.Atoi(string(buf))
	if err != nil {
		t.Error(err)
		return
	}
	if lv != rv {
		t.Errorf("Layouter[%T] failed! expected: %v, got: %v", layout, rv, lv)
	}
}

func compareStr(t *testing.T, layout Layouter, time time.Time, rv string) {
	var buf []byte
	layout.layout(&buf, DebugLevel, "", time, "", 0)
	if strings.Compare(string(buf), rv) != 0 {
		t.Errorf("Layouter[%T] failed! expected: %q, got: %q", layout, rv, string(buf))
	}
}

func TestLayouter(t *testing.T) {
	now := time.Now()
	compareInt(t, &logouterYear{}, now, now.Year())
	compareInt(t, &logouterMonth{}, now, int(now.Month()))
	compareInt(t, &logouterDay{}, now, now.Day())
	compareInt(t, &logouterHour{}, now, now.Hour())
	compareInt(t, &logouterMinute{}, now, now.Minute())
	compareInt(t, &logouterSecond{}, now, now.Second())

	dateStr := fmt.Sprintf("%04d", int(now.Year())) + "/" + fmt.Sprintf("%02d", int(now.Month())) + "/" + fmt.Sprintf("%02d", int(now.Day()))
	timeStr := fmt.Sprintf("%02d", int(now.Hour())) + ":" + fmt.Sprintf("%02d", int(now.Minute())) + ":" + fmt.Sprintf("%02d", int(now.Second()))
	compareStr(t, &layouterDate{}, now, dateStr)
	compareStr(t, &layouterTime{}, now, timeStr)
}

func TestDefaultLogLayout(t *testing.T) {
	lis := &strLogListener{}
	logger := New(InfoLevel, lis, "")
	logMsg := "test log"
	logger.Info(logMsg)
	now := time.Now()
	dateStr := fmt.Sprintf("%04d", int(now.Year())) + "/" + fmt.Sprintf("%02d", int(now.Month())) + "/" + fmt.Sprintf("%02d", int(now.Day()))
	timeStr := fmt.Sprintf("%02d", int(now.Hour())) + ":" + fmt.Sprintf("%02d", int(now.Minute())) + ":" + fmt.Sprintf("%02d", int(now.Second()))
	expectedLog := "[INFO] " + dateStr + " " + timeStr + " " + logMsg + "\n"
	if strings.Compare(lis.log, expectedLog) != 0 {
		t.Errorf("'%v' expected, got '%v'", expectedLog, lis.log)
	}
}

func TestCustomLogLayout(t *testing.T) {
	lis := &strLogListener{}
	logger := New(InfoLevel, lis, "!!!%L - %D -- %T---%l!!!")
	logMsg := "test log"
	logger.Info(logMsg)
	now := time.Now()
	dateStr := fmt.Sprintf("%04d", int(now.Year())) + "/" + fmt.Sprintf("%02d", int(now.Month())) + "/" + fmt.Sprintf("%02d", int(now.Day()))
	timeStr := fmt.Sprintf("%02d", int(now.Hour())) + ":" + fmt.Sprintf("%02d", int(now.Minute())) + ":" + fmt.Sprintf("%02d", int(now.Second()))
	expectedLog := "!!![INFO] - " + dateStr + " -- " + timeStr + "---" + logMsg + "!!!\n"
	if strings.Compare(lis.log, expectedLog) != 0 {
		t.Errorf("'%v' expected, got '%v'", expectedLog, lis.log)
	}
}

func TestLogLevel(t *testing.T) {
	lis := &strLogListener{}
	logger := New(InfoLevel, lis, "!!!%L - %D -- %T---%l!!!")
	logMsg := "test log"
	logger.Debug(logMsg)
	if len(lis.log) != 0 {
		t.Errorf("'DEBUG' level log should not be output on 'INFO' log level.")
	}
}
