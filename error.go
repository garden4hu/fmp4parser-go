package fmp4parser

import "errors"

var (
	ErrNoEnoughData = errors.New("no enough data")
	ErrOutOfRange	= errors.New("out of range when set pos")
	ErrUnsupportedSampleEntry = errors.New("unsupported sample entry ")
	ErrIncompleteCryptoBox = errors.New("incomplete box about enca/encv ")
)

var (
	ErrEof           = errors.New("EOF")
	ErrUnexpectedEof = errors.New("unexpected EOF")
)
