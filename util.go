package fmp4parser

import (
	"io"
	"log"
)

// The status parameter represents the top level box currently being processed
const (
	StateFtyp int = iota // value --> 0
	StateMOOV            // value --> 1
	StateMOOF            // value --> 2
)

var logs *logObj

// logger provide a simpler log writer
type logObj struct {
	err   *log.Logger
	warn  *log.Logger
	info  *log.Logger
	debug *log.Logger
}

func newLogger(w io.Writer) *logObj {
	logger := new(logObj)
	logger.err = log.New(w, "[E] [fmp4parser]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.warn = log.New(w, "[W] [fmp4parser]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.info = log.New(w, "[I] [fmp4parser]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	logger.debug = log.New(w, "[D] [fmp4parser]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	return logger
}
