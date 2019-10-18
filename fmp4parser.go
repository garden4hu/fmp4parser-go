package fmp4parser

import (
	"io"
	"os"
)

// Fmp4Parser is a object type for fmp4parser
type Fmp4Parser struct {
	obj *parser
}

// NewFmp4Parser return a pointer to a variable of Fmp4Parser
func NewFmp4Parser() *Fmp4Parser {
	return &Fmp4Parser{obj: NewParser()}
}

// Init do some initialization work. Invoker should provide an variable of io.Writer
func (h *Fmp4Parser) Init(logWriter io.Writer){
	if logWriter == nil {
		logs = newLogger(os.Stdout)
	}else {
		logs = newLogger(logWriter)
	}
}

func (h *Fmp4Parser) Process(rawData []byte) error {
	if len(rawData) == 0 {
		return ErrNoEnoughData
	}
	h.obj.Append(rawData)
	// TODO
	return nil
}
