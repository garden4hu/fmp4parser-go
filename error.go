package fmp4parser

import "errors"

var (
	ErrNoEnoughData = errors.New("no enough data")
	ErrOutOfRange	= errors.New("out of range when set pos")
)

var (
	ErrEof           = errors.New("EOF")
	ErrUnexpectedEof = errors.New("unexpected EOF")
)
