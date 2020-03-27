package fmp4parser

import "errors"

var (
	ErrNoEnoughData           = errors.New("no enough data")
	ErrOutOfRange             = errors.New("out of range when set pos")
	ErrRequestTooLarge        = errors.New("request data is too large")
	ErrIncompleteBox          = errors.New("incomplete box")
	ErrUnsupportedSampleEntry = errors.New("unsupported sample entry ")
	ErrIncompleteCryptoBox    = errors.New("incomplete box of enca/encv ")
)

var (
	ErrEof           = errors.New("EOF")
	ErrUnexpectedEof = errors.New("unexpected EOF")
)
