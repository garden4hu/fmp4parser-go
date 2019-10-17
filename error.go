package fmp4parser

import "errors"

var (
	ErrNoEnoughData = errors.New("no enough data")
)

var (
	ErrEof           = errors.New("EOF")
	ErrUnexpectedEof = errors.New("unexpected EOF")
)
