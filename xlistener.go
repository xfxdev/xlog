package xlog

import (
	"bufio"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// A Listener simple typed of io.Writer
type Listener io.WriteCloser

// W2FileListener use to output log to file.
type W2FileListener struct {
	w *bufio.Writer
	f *os.File
}

// Write is equivalent to os.File.Write.
func (l *W2FileListener) Write(p []byte) (n int, err error) {
	return l.w.Write(p)
}

// Close is equivalent to os.File.Close.
func (l *W2FileListener) Close() error {
	if l != nil && l.f != nil {
		err := l.w.Flush()
		if err2 := l.f.Close(); err2 != nil {
			if err == nil {
				err = err2
			}
		}
		return err
	}
	return os.ErrInvalid
}

// NewW2FileListener creates a new W2FileListener.
// If filePath is empty, then will create file at appPath/Log/
func NewW2FileListener(filePath string) (*W2FileListener, error) {
	if len(filePath) == 0 {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return nil, err
		}

		baseName := filepath.Base(os.Args[0])
		appName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		timeStr := time.Now().Format("2006_01_02")
		filePath = path.Join(appDir, "log", appName+"_"+timeStr+".log")
	}
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	lis := &W2FileListener{
		f: f,
		w: bufio.NewWriter(f),
	}

	return lis, nil
}
