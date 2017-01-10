# xlog
plugin architecture and flexible log system for golang

[![Build Status](https://travis-ci.org/xfxdev/xlog.svg?branch=master)](https://travis-ci.org/xfxdev/xlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/xfxdev/xlog)](https://goreportcard.com/report/github.com/xfxdev/xlog)
[![GoDoc](https://godoc.org/github.com/xfxdev/xlog?status.svg)](https://godoc.org/github.com/xfxdev/xlog)

Installation
================
~~~
go get github.com/xfxdev/xlog
~~~

Usage
================
~~~
import (
    "github.com/xfxdev/xlog"
)

strLogLevel := "INFO"  // also maybe read from config.
logLevel, suc := xlog.ParseLevel(strLogLevel)
if suc == false {
    // failed to parse log level, will use the default level[INFO] instead."
}
xlog.SetLevel(logLevel)

// write log to file.
w2f, err := xlog.NewW2FileListener("")
if err != nil {
    xlog.Fatal(err)
} else {
    xlog.AddListener(w2f)
}

xlog.Info("server start...")
xlog.Debugf("ip : %v", "127.0.0.1")
~~~

Features
================
###Level logging
~~~
// xlog provide 6 logging levels.
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)
~~~

you can call 'SetLevel' to change the log level. All the logs which level <= your set will be output.
for example:
~~~
// all logs will be output because DEBUG is the max level.
xlog.SetLevel(xlog.DebugLevel) 

// only Panic/Fatal/Error logs will be output because WARN/INFO/DEBUG > ERROR.
xlog.SetLevel(xlog.ErrorLevel)
~~~

###Custom Log Layout
xlog provide a flexible log layout system to custom the style of log output.
What you need to do is just set the layout flags.
~~~
xlog.SetLayout('layout flags...')
~~~
xlog provide some builtin layout:
~~~
//   %y : year
//   %M : month
//   %d : day
//   %h : hour
//   %m : min
//   %s : second
//   %l : log msg
//   %L : log level
//   %F : file			eg: /a/b/c/d.go
//   %f : short file	eg: d.go
//   %i : line
//   %D : %y/%M/%d
//   %T : %h:%m:%s
~~~
You can use a combination of them, for example:
~~~
// this mean every log message will have a '[level] year/month/day hour:min:sec' perfix, eg:
xlog.SetLayout("%L %D %T %l")

// outputs:
[INFO] 2016/01/01 13:27:07 net start...
[WARN] 2016/01/01 13:28:00 ...
[DEBUG] 2016/01/01 13:28:00 accept...

// add filename and line to log message.
xlog.SetLayout("%L %D %T [%f(%i)] %l")

// outputs:
[INFO] 2016/01/01 13:27:07 [test.go:(72)] net start...
[WARN] 2016/01/01 13:28:00 [test.go:(100)] ...
[DEBUG] 2016/01/01 13:28:00 [test.go:(128)] accept...
~~~
You can use any form of combination, even meaningless thing, such as more spaces, arbitrary symbols:
~~~
xlog.SetLayout("hahaha%L | %D   %T [ABC] %l [i'm after hahaha]")

// outputs:
// notice the prefix 'hahaha', the spaces in the middle, and the suffix '[i'm after hahaha]'
hahaha[INFO] | 2017/01/07     14:09:47 [ABC] net start... [i'm after hahaha]
hahaha[WARN] | 2017/01/07     14:09:47 [ABC] ... [i'm after hahaha]
hahaha[DEBUG] | 2017/01/07     14:09:47 [ABC] accept... [i'm after hahaha]
~~~
NOTICE: If you doesn't call 'SetLayout', xlog will use '%L %D %T %l' by default.

###Output a log to different targets
xlog use listener system to output the log message.
~~~
// A Listener simple typed of io.Writer
type Listener io.Writer
~~~
A logger can have multiple listeners, xlog have 2 builtin listener, which are os.Stderr and W2FileListener.
xlog will output log to os.Stderr by default, but you can add W2FileListener to output the log to file.
~~~
w2f, err := xlog.NewW2FileListener("logfilePath...")
if err != nil {
    xlog.Fatal(err)
} else {
    xlog.AddListener(w2f)
}
~~~
In 'NewW2FileListener' function, xlog will use the 'logfilePath' to create log file,
so please makesure your path is correct.
~~~
os.MkdirAll(filepath.Dir(logfilePath), os.ModePerm)
~~~
Also, you can give a empty path to NewW2FileListener(""), this will create log file by simple rule. for example:
If your app current path is 'a/b/c', and your app name is 'testapp'
Then the log file will be 'a/b/c/log/testapp_2016_01_01.log'
~~~
w2f, err := xlog.NewW2FileListener("")  // create log file at : 'a/b/c/log/test_2016_01_01.log'
if err != nil {
    xlog.Fatal(err)
}
xlog.AddListener(w2f)
~~~
You can create new listener according you need, just implement the io.Writer interface.
~~~
type Writer interface {
	Write(p []byte) (n int, err error)
}
~~~
###Thread safety
By default xlog is protected by mutex, so you can output logs in multiple goroutines.
