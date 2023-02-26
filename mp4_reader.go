package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// mp4Reader handles all things related to the buffer.
// It receives an io.ReadSeeker from invoker and reads un-parsed
// data from the latter.
type mp4Reader struct {
	readSeeker io.ReadSeeker
	b          []byte // processing bytes
	a          *atom  // processing atom
	startPos   int64
	endPos     int64
}

func newMp4Reader(i io.ReadSeeker) *mp4Reader {
	return &mp4Reader{
		readSeeker: i,
	}
}

func (p *mp4Reader) GetAtomPosition() int64 {
	return p.startPos
}

// CheckAtomParseEnd checks whether the currently
// processed data is outside the buffer range of the current atom.
func (p *mp4Reader) CheckAtomParseEnd() bool {
	n, _ := p.readSeeker.Seek(0, io.SeekCurrent)
	if n >= p.endPos {
		return true
	}
	return false
}

// PeekAtomHeader will try to peek the buffer and get the atom's
// type/size information without moving the file pointer.
func (p *mp4Reader) PeekAtomHeader() (a *atom, err error) {
	startPos, _ := p.readSeeker.Seek(0, io.SeekCurrent)
	a, err = p.ReadAtomHeader()
	p.readSeeker.Seek(startPos, io.SeekStart)
	return a, nil
}

// ReadAtomHeader will read the next atom's header if no error occur.
// If it returns without error, the new atom will replace the previous one.
// And the previous atomReader will be invalid.
func (p *mp4Reader) ReadAtomHeader() (a *atom, err error) {
	readInt := func(b []byte) uint32 {
		return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}

	if err != nil {
		return nil, err
	}
	header := make([]byte, 8)
	n, e := p.Read(header)
	if n != 8 {
		return nil, e
	}
	a = new(atom)
	a.bodySize = int64(readInt(header[:4]))
	a.atomType = readInt(header[4:8])
	if a.bodySize == 1 {
		realSize := make([]byte, 8)
		n, e = p.Read(realSize)
		if n != 8 {
			return nil, e
		}
		a.headerSize = 16
		a.bodySize = int64(readInt(realSize[:4]))<<32 | int64(readInt(realSize[4:8])) - 16
		header = append(header, realSize...)
	} else {
		a.bodySize -= 8
		a.headerSize = 8
	}

	p.a = a
	p.b = nil
	p.startPos, _ = p.readSeeker.Seek(0, io.SeekCurrent)
	p.endPos = p.startPos + a.bodySize
	p.startPos -= int64(a.headerSize)
	return a, nil
}

func (p *mp4Reader) ReadAtomData() error {
	p.b = make([]byte, p.a.bodySize)
	n, _ := p.readSeeker.Read(p.b)
	if int64(n) != p.a.bodySize {
		p.readSeeker.Seek(int64(-n), io.SeekCurrent)
		return ErrNoEnoughData
	}
	return nil
}

/*
1. first time read the atom: p.a == nil && p.b == nil
2. has tried to read the atom but failed, and retry read
3.

*/

// GetAtom return an atomReader if no error encountered.
// If the ReadSeeker failed to read atom's size of buffer, it will
// return error and restore the read pointer.
func (p *mp4Reader) GetAtom() (*atomReader, error) {
	if _, err := p.ReadAtomHeader(); err != nil {
		return nil, nil
	}
	if err := p.ReadAtomData(); err != nil {
		return nil, err
	}
	return newAtomReader(p.b, p.a), nil
}

// SkipCurrentAtom will skip the following atom. It must be called
// in the boundary of the atoms.
func (p *mp4Reader) SkipCurrentAtom() (err error) {
	currentPos, _ := p.readSeeker.Seek(0, io.SeekCurrent)
	_, err = p.readSeeker.Seek(p.a.Size(), io.SeekCurrent)
	if err == nil {
		return nil
	}
	_, _ = p.readSeeker.Seek(currentPos, io.SeekStart) // restore the reader
	return ErrOperationWithDraw
}

// use it to read sample in "mdat"
func (p *mp4Reader) Read(b []byte) (n int, err error) {
	return p.readSeeker.Read(b)
}

func (p *mp4Reader) getReaderPosition() int64 {
	n, _ := p.readSeeker.Seek(0, io.SeekCurrent)
	return n
}

// Peek will read at most len(b) bytes without moving the reading pointer.
// Note that the no-nil error should be processed.
func (p *mp4Reader) Peek(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	n, err = p.readSeeker.Read(b)
	p.readSeeker.Seek(int64(-n), io.SeekCurrent)
	return n, err
}

// bitReader wraps an io.Reader and provides the ability to read values,
// bit-by-bit, from it. Its Read* methods don't return the usual error
// because the error handling was verbose. Instead, any error is kept and can
// be checked afterwards.
// modify from https://golang.org/src/compress/bzip2/bit_reader.go
type bitReader struct {
	r    io.ByteReader
	n    uint64
	bits uint
	err  error
}

func newBitReader(r io.Reader) bitReader {
	byteReader, ok := r.(io.ByteReader)
	if !ok {
		byteReader = bufio.NewReader(r)
	}
	return bitReader{r: byteReader}
}

func newBitReaderFromSlice(src []byte) bitReader {
	return newBitReader(bytes.NewReader(src))
}

// ReadBitsLE64 when bits <= 64
func (br *bitReader) ReadBitsLE64(bits uint) (n uint64) {
	for bits > br.bits {
		b, err := br.r.ReadByte()
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			br.err = err
			return 0
		}
		br.n <<= 8
		br.n |= uint64(b)
		br.bits += 8
	}
	n = (br.n >> (br.bits - bits)) & ((1 << bits) - 1)
	br.bits -= bits
	return
}

// ReadBitsLE32 only when bits <= 32
func (br *bitReader) ReadBitsLE32(bits uint) (n uint32) {
	n64 := br.ReadBitsLE64(bits)
	return uint32(n64)
}

// ReadBitsLE8 read less(equal) than 8 bits
func (br *bitReader) ReadBitsLE8(bits uint) (n uint8) {
	return uint8(br.ReadBitsLE64(bits))
}

// ReadBitsLE16 read less(equal) than 16 bits
func (br *bitReader) ReadBitsLE16(bits uint) (n uint16) {
	return uint16(br.ReadBitsLE64(bits))
}

func (br *bitReader) ReadBool() bool {
	n := br.ReadBitsLE32(1)
	return n != 0
}

func (br *bitReader) Err() error {
	return br.err
}

func int2String(n uint32) string {
	return fmt.Sprintf("%c%c%c%c", uint8(n>>24), uint8(n>>16), uint8(n>>8), uint8(n))
}
func string2int(s string) uint32 {
	if len(s) != 4 {
		logE.Printf("string2int, the length of %s is not 4", s)
	}
	b := []byte(s)
	b = b[0:4]
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
