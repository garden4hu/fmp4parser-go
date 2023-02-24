package main

import "errors"

var (
	ErrInvalidParam           = errors.New("invalid parameter")
	ErrNoEnoughData           = errors.New("no enough Data")
	ErrAtomNotFound           = errors.New("atom not found")
	ErrOutOfRange             = errors.New("out of range when set pos")
	ErrRequestTooLarge        = errors.New("request Data is too large")
	ErrIncompleteBox          = errors.New("incomplete box")
	ErrUnsupportedSampleEntry = errors.New("unsupported sample entry ")

	ErrInvalidAtomSize = errors.New("atom's size is invalid")
	ErrNoMoreAtom      = errors.New("there is no  more atom")
	ErrInvalidAtom     = errors.New("the atom is bad")

	ErrUnsupportedAtomType = errors.New("the atom isn't supported yet")

	ErrMoovNotParsed                        = errors.New("moov atom(movie header) is not parsed yet")
	ErrIncompleteCryptoBox                  = errors.New("incomplete box of protectedInfo/protectedInfo ")
	ErrUnsupportedSampleGroupType           = errors.New("unsupported sampleToGroup type")
	ErrUnsupportedVariableSampleGroupLength = errors.New("sampleToGroup entry is variable, unsupported currently")
	ErrInvalidLengthOfSampleGroup           = errors.New("length individual sampleToGroup entry is invalid")
	ErrInvalidLengthOfIVInSampleGroup       = errors.New("in cenc sample group entry, the length of (const)IV is not 8 or 16")

	ErrNotFoundTrak = errors.New("not found the trak information in moov")
	ErrNoImplement  = errors.New("function parse has not been implement")
)

var (
	ErrEof           = errors.New("EOF")
	ErrUnexpectedEof = errors.New("unexpected EOF")
	ErrTooLarge      = errors.New("file is too large")
)
