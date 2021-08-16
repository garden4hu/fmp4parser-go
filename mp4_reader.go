package fmp4parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// mp4Reader handles all things related to the buffer.
// It receives an io.Reader from fmp4parser API and reads un-parsed
// data from the latter.
type mp4Reader struct {
	r io.Reader // IO outside
	b *mp4Buffer
}

// newMp4Reader return an mp4Reader
func newMp4Reader(r io.Reader) *mp4Reader {
	return &mp4Reader{
		r: r,
		b: new(mp4Buffer),
	}
}

func (p *mp4Reader) MoveTo() error {
	return nil
}

// Notice: blow functions will not check the availability of underlying mp4Buffer.
// When using these functions, the mp4Buffer is always considered meeting the needs.
// ReadUnsignedByte, ReadSignedByte, Read2, Read2S, Read4, Read4S, Read8, Read8S,
// ReadAtomHeader, ReadVersionFlags, FindSubAtom

// ReadUnsignedByte read 1 byte and return uint8
func (p *mp4Reader) ReadUnsignedByte() uint8 {
	c, _ := p.b.ReadByte()
	return c
}

// ReadSignedByte read 1 byte and return int8
func (p *mp4Reader) ReadSignedByte() int8 {
	c, _ := p.b.ReadByte()
	return int8(c)
}

// Read2 read 2 byte from mp4Buffer slice and return the bitwise-integer if no error
func (p *mp4Reader) Read2() uint16 {
	h := p.ReadUnsignedByte()
	l := p.ReadUnsignedByte()
	return uint16(h)<<8 | uint16(l)
}

// Read2S read 2 byte from mp4Buffer slice and return a signed integer(16bits)if no error
func (p *mp4Reader) Read2S() int16 {
	h := p.ReadSignedByte()
	l := p.ReadUnsignedByte()
	return int16(h)<<8 | int16(l)
}

// Read4 read 4 bytes from mp4Buffer slice and return an unsigned integer(32bits) if no error
func (p *mp4Reader) Read4() uint32 {
	h := p.Read2()
	l := p.Read2()
	return uint32(h)<<16 | uint32(l)
}

// Read4S read 4 bytes from mp4Buffer slice and return a signed integer(32bits) if no error
func (p *mp4Reader) Read4S() int32 {
	h := p.Read2S()
	l := p.Read2()
	return int32(h)<<16 | int32(l)
}

// Read8 read 8 bytes from mp4Buffer slice and return an unsigned integer(64bits) if no error
func (p *mp4Reader) Read8() uint64 {
	h := p.Read4()
	l := p.Read4()
	return uint64(h)<<32 | uint64(l)
}

// Read8S read 8 bytes from mp4Buffer slice and return a signed integer(64bits) if no error
func (p *mp4Reader) Read8S() int64 {
	h := p.Read4S()
	l := p.Read4()
	return int64(h)<<32 | int64(l)
}

// ReadAtomHeader parse the atom from the mp4Buffer
func (p *mp4Reader) ReadAtomHeader() *atom {
	a := new(atom)
	fullAtomSize := int64(p.Read4())
	a.atomType = p.Read4()
	if fullAtomSize == 1 { // full box
		fullAtomSize = int64(p.Read8())
		a.atomHeaderSize = 16
	} else {
		a.atomHeaderSize = 8
	}
	a.atomSize = fullAtomSize - int64(a.atomHeaderSize)
	return a
}

// PeakAtomHeader parse the atom from the mp4Buffer without moving the buffer pointer
func (p *mp4Reader) PeakAtomHeader() *atom {
	a := new(atom)
	fullAtomSize := int64(p.Read4())
	a.atomType = p.Read4()
	if fullAtomSize == 1 { // full box
		fullAtomSize = int64(p.Read8())
		a.atomHeaderSize = 16
	} else {
		a.atomHeaderSize = 8
	}
	a.atomSize = fullAtomSize - int64(a.atomHeaderSize)
	return a
}

// ReadVersionFlags return the Version of the box and the flags of the box
func (p *mp4Reader) ReadVersionFlags() (uint8, uint32) {
	n := p.Read4()
	return uint8(n >> 24 & 0xFF), n & 0x00FFFFFF
}

// FindSubAtom return the size of the box in current atom if error is nil.
// The reading pointer points to the starting of sub-atom's body mp4Buffer if error is nil.
// Otherwise, the reading pointer keeps no change if error is ErrAtomNotFound
// Usually use this function to find the non-top Level box type, such as 'mvhd', 'esds' etc.
// Searching the top Level box type is not support because it will coast massive memory.
func (p *mp4Reader) FindSubAtom(boxType uint32, scope int) (int64, error) {
	if scope < 8 {
		return 0, ErrAtomNotFound
	}
	left := scope
	curBuffLen := p.Len()
	for left >= 8 {
		a := p.ReadAtomHeader()
		if a.atomType == boxType {
			return a.atomSize, nil
		} else {
			_ = p.Move(int(a.atomSize))
			left -= int(a.Size())
		}
	}
	p.Move(p.Len() - curBuffLen)
	return -1, ErrAtomNotFound
}

// ReadBytes read n bytes from mp4Buffer and return the n-bytes mp4Buffer
// and the size of n-bytes mp4Buffer if error is nil
func (p *mp4Reader) ReadBytes(b []byte) (int, error) {
	if b == nil || len(b) == 0 {
		return 0, nil
	}
	_, err := p.b.Peek(b)
	if err != nil {
		return 0, fmt.Errorf("%w ReadBytes", err)
	}
	p.b.Move(len(b))
	return len(b), nil
}

// Reset will reset the mp4Buffer to reuse the memory
func (p *mp4Reader) Reset() {
	p.b.Reset()
}

// Len return the size of un-read mp4Buffer
func (p *mp4Reader) Len() int {
	return p.b.Len()
}

// Move wrap mp4Buffer.Move
func (p *mp4Reader) Move(n int) bool {
	return p.b.Move(n)
}

// Peek return the next n byte unread Data without changing the internal status.
// if the read size isn't len(b), return error
func (p *mp4Reader) Peek(b []byte) (int, error) {
	if p.Len() < len(b) {
		err := p.Append(len(b) - p.Len())
		if err != nil {
			return 0, err
		}
	}
	return p.b.Peek(b)
}

// Append is using for appending a slice of byte to current p.buf
func (p *mp4Reader) Append(n int) error {
	_, e := p.b.ReadBytesFromAtLeast(p.r, n)
	return e
}

// AppendAtomHeader try to append the mp4Buffer of next atom's header.
// Use this function when the coming atom size is unknown.
func (p *mp4Reader) AppendAtomHeader() error {
	if p.b.Len() < 8 {
		err := p.Append(8 - p.b.Len())
		if err != nil {
			return err
		}
	}
	atomSize := p.Read4()
	p.Move(-4)
	if atomSize == 1 {
		err := p.Append(8)
		if err != nil {
			return err
		}
	}
	return nil
}

// AppendAtomBody try to append the body of the atom form the information of atom.
func (p *mp4Reader) AppendAtomBody(a *atom) error {
	return p.Append(int(a.atomSize))
}

// AppendAtom return the next entire-atom's mp4Buffer
func (p *mp4Reader) AppendAtom() error {
	e := p.AppendAtomHeader()
	if e != nil {
		return e
	}
	a := p.ReadAtomHeader()
	p.Move(int(-a.atomHeaderSize))
	e = p.AppendAtomBody(a)
	if e != nil {
		return e
	}
	return nil
}

// ResetReader reset the reader
func (p *mp4Reader) ResetReader(r io.Reader) {
	p.r = r
}

// Position return the distance between br's pointer and the beginning of mp4Buffer
func (p *mp4Reader) Position() int64 {
	return 0
}

func (p *mp4Reader) HasMoreData() bool {
	if p.Len() <= 8 || p.Append(8) != nil {
		return false
	}
	return true
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
