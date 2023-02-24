package main

import (
	"io"
	"log"
)

var (
	logE *log.Logger // error
	logW *log.Logger // warn
	logI *log.Logger // info
	logD *log.Logger // debug
)

func newLog(out io.Writer) {
	logE = log.New(out, "[E] [fmp4parser] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
	logW = log.New(out, "[W] [fmp4parser] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
	logI = log.New(out, "[I] [fmp4parser] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
	logD = log.New(out, "[D] [fmp4parser] ", log.Ldate|log.Ltime|log.Lmicroseconds|log.LstdFlags|log.Lshortfile)
}
