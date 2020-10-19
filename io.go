package fmp4parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type deMuxReader struct {
	readSeeker io.ReadSeeker // IO br
	buf        []byte        // buffer to cache the un-parsed data
	br         *bytes.Reader // local buffer reader
}

func NewDeMuxReader(r io.ReadSeeker) *deMuxReader {
	de := deMuxReader{readSeeker: r}
	de.br = bytes.NewReader(de.buf)
	return &de
}

// ReadUnsignedByte read 1 byte and return uint8
func (p *deMuxReader) ReadUnsignedByte() uint8 {
	c, _ := p.br.ReadByte()
	return c
}

// ReadSignedByte read 1 byte and return int8
func (p *deMuxReader) ReadSignedByte() int8 {
	c, _ := p.br.ReadByte()
	return int8(c)
}

// Read2 read 2 byte from buffer slice and return the bitwise-integer if no error
func (p *deMuxReader) Read2() uint16 {
	byteHeight := p.ReadUnsignedByte()
	byteLow := p.ReadUnsignedByte()
	return uint16(byteHeight)<<8 | uint16(byteLow)
}

// Read2S read 2 byte from buffer slice and return a signed integer(16bits)if no error
func (p *deMuxReader) Read2S() int16 {
	byteHeight := p.ReadSignedByte()
	byteLow := p.ReadUnsignedByte()
	return int16(byteHeight)<<8 | int16(byteLow)
}

// Read4 read 4 bytes from buffer slice and return an unsigned integer(32bits) if no error
func (p *deMuxReader) Read4() uint32 {
	bytesInt := make([]byte, 4, 4)
	// _, errLog := io.ReadAtLeast(&p.br, bytesInt, 4)
	_, _ = p.br.Read(bytesInt)
	return uint32(bytesInt[0])<<24 | uint32(bytesInt[1])<<16 | uint32(bytesInt[2])<<8 | uint32(bytesInt[3])
}

// Read4S read 4 bytes from buffer slice and return a signed integer(32bits) if no error
func (p *deMuxReader) Read4S() int32 {
	bytesInt := make([]byte, 4, 4)
	// _, errLog := io.ReadAtLeast(&p.br, bytesInt, 4)
	_, _ = p.br.Read(bytesInt)
	return int32(bytesInt[0])<<24 | int32(bytesInt[1])<<16 | int32(bytesInt[2])<<8 | int32(bytesInt[3])
}

// Read8 read 8 bytes from buffer slice and return an unsigned integer(64bits) if no error
func (p *deMuxReader) Read8() uint64 {
	byteLong := make([]byte, 8, 8)
	_, _ = p.br.Read(byteLong)
	high := uint32(byteLong[0])<<24 | uint32(byteLong[1])<<16 | uint32(byteLong[2])<<8 | uint32(byteLong[3])
	low := uint32(byteLong[4])<<24 | uint32(byteLong[5])<<16 | uint32(byteLong[6])<<8 | uint32(byteLong[7])
	return uint64(high)<<32 | uint64(low)
}

// Read8S read 8 bytes from buffer slice and return a signed integer(64bits) if no error
func (p *deMuxReader) Read8S() int64 {
	byteLong := make([]byte, 8, 8)
	_, _ = p.br.Read(byteLong)
	high := int32(byteLong[0])<<24 | int32(byteLong[1])<<16 | int32(byteLong[2])<<8 | int32(byteLong[3])
	low := int32(byteLong[4])<<24 | int32(byteLong[5])<<16 | int32(byteLong[6])<<8 | int32(byteLong[7])
	return int64(high)<<32 | int64(low)
}

// ReadBytes read n bytes from buffer and return the n-bytes buffer
// and the size of n-bytes buffer if error is nil
func (p *deMuxReader) ReadBytes(n int) ([]byte, int, error) {
	if n <= 0 {
		return nil, -1, io.ErrUnexpectedEOF
	}
	byteArr := make([]byte, n, n)
	nRet, _ := p.br.Read(byteArr)
	return byteArr, nRet, nil
}

// ReadAtomHeader return the atom's type and size if no error encountered
// the atom.size doesn't contain the header size of the atom
func (p *deMuxReader) ReadAtomHeader() *atom {
	if p.br.Len() < 8 {
		return nil
	}
	var a atom
	fullAtomSize := int64(p.Read4())
	a.atomType = p.Read4()
	if fullAtomSize == 1 { // full box
		fullAtomSize = int64(p.Read8())
		a.atomHeaderSize = 16
	} else {
		a.atomHeaderSize = 8
	}
	a.atomSize = fullAtomSize - int64(a.atomHeaderSize)
	return &a
}

// ReadVersionFlags return the Version of the box and the flags of the box
func (p *deMuxReader) ReadVersionFlags() (uint8, uint32) {
	n := p.Read4()
	return uint8(n >> 24 & 0xFF), n & 0x00FFFFFF
}

// DiscardUsedData drop the used Data of buffer [remove form 0 to p.index]
func (p *deMuxReader) DiscardUsedData() {
	currentIndex, _ := p.br.Seek(0, io.SeekCurrent)
	copy(p.buf[0:], p.buf[currentIndex:])
	for k, n := int64(len(p.buf))-currentIndex, int64(len(p.buf)); k < n; k++ {
		p.buf[k] = 0
	}
	p.buf = p.buf[:int64(len(p.buf))-currentIndex]
	p.br.Reset(p.buf)
}

// Reset reset the br pointer to the begin of slice
func (p *deMuxReader) Reset() {
	p.br.Reset(p.buf)
}

// Move wrap the io.Seeker without error check
func (p *deMuxReader) Move(size int64) {
	_, _ = p.br.Seek(size, io.SeekCurrent)
}

// MoveTo try to move the absolute position of the slice
func (p *deMuxReader) MoveTo(pos int64) error {
	if pos > int64(len(p.buf)) {
		return ErrOutOfRange
	}
	_, err := p.br.Seek(pos, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

// Peek return the next n byte unread Data without change the br's index.
// if the condition is not met, return error
func (p *deMuxReader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrOutOfRange
	}
	if n > p.br.Len() {
		return nil, ErrRequestTooLarge
	}
	currentIndex, _ := p.br.Seek(0, io.SeekCurrent)
	return p.buf[currentIndex : currentIndex+int64(n)], nil
}

// FindAtom return the size of the specific type of iso box if error equal nil
// usually use this function to find the top Level box type, such as 'ftyp', 'moov' etc.
func (p *deMuxReader) FindAtom(boxType uint32) (int64, error) {
	err := errors.New("")
	for {
		a := p.ReadAtomHeader()
		if a.atomType == boxType {
			return a.atomSize, nil
		} else {
			_, err := p.br.Seek(a.atomSize, io.SeekCurrent)
			if err != nil {
				goto T
			}
		}
	}
T:
	return -1, err
}

// FindAtomWithinScope return the size of the box in the specific interval if error is nil
// Usually use this function to find the non-top Level box type, such as 'mvhd', 'esds' etc.
// Searching the top Level box type is also support.
func (p *deMuxReader) FindAtomWithinScope(boxType uint32, scope int64) (int64, error) {
	if scope < 8 {
		return -1, ErrAtomNotFound
	}
	leftSize := scope
	startPosition := p.Position()
	for leftSize > 8 {
		a := p.ReadAtomHeader()
		// logD.Print(a)
		//	break
		if a.atomType == boxType {
			return a.atomSize, nil
		} else {
			_ = p.MoveTo(p.Position() + a.atomSize)
			leftSize -= p.Position() - startPosition
		}
	}
	return -1, ErrAtomNotFound
}

// Append is using for appending a slice of byte to current p.buf
func (p *deMuxReader) Append(n int64) error {
	buff := make([]byte, n)
	nRet, err := p.readSeeker.Read(buff)
	if err != nil {
		logE.Print("", err)
		return err
	} else if int64(nRet) < n {
		_, _ = p.readSeeker.Seek(int64(-nRet), io.SeekEnd)
		logE.Print("failed to read enough data ,error type is:", ErrNoEnoughData)
		return ErrNoEnoughData
	}
	cur := p.Position()
	p.buf = append(p.buf, buff...)
	p.br.Reset(p.buf)
	_ = p.MoveTo(cur)
	return nil
}

// Len return the size of un-read buffer
func (p *deMuxReader) Len() int64 {
	return int64(p.br.Len())
}

// AppendNextAtom return the next entire-atom's buffer
func (p *deMuxReader) AppendNextAtom() error {
	if p.Len() < 8 {
		if err := p.Append(8); err != nil {
			return err
		}
	}
	currentPosition := p.Position()
	a := p.ReadAtomHeader()
	err := p.Append(a.atomSize)
	if err != nil {
		logE.Printf("failed to Append atom:%s's body data", a.Type())
		_ = p.MoveTo(currentPosition)
		return err
	}
	_ = p.MoveTo(currentPosition)
	return err
}

// Position return the distance between br's pointer and the begin of buffer
func (p *deMuxReader) Position() int64 {
	index, _ := p.br.Seek(0, io.SeekCurrent)
	return index
}

func (p *deMuxReader) HasMoreData() bool {
	if p.Len() <= 8 || p.Append(8) != nil {
		return false
	}
	return true
}

// ResetReader for the io.ReadSeeker
func (p *deMuxReader) ResetReader(r io.ReadSeeker) {
	p.readSeeker = r
	p.Move(p.Len())
	p.DiscardUsedData() // clear cached data
	p.Reset()
}

// Skip try to skip n bytes un-read buffer
func (p *deMuxReader) Skip(offset int64) error {
	if offset < 0 {
		return ErrInvalidParam
	}
	if p.Len() < offset {
		currentPosition, _ := p.readSeeker.Seek(0, io.SeekCurrent)
		_, err := p.readSeeker.Seek(offset-p.Len(), io.SeekCurrent)
		if err != nil {
			_, _ = p.readSeeker.Seek(currentPosition, io.SeekStart)
			return io.ErrUnexpectedEOF
		}
		p.Clear()
	} else {
		p.Move(offset)
		p.DiscardUsedData()
	}
	return nil
}

// Clear remove the data in the buffer
func (p *deMuxReader) Clear() {
	p.Move(p.Len())
	p.DiscardUsedData()
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

// use ReadBitsLE64 when bits <= 64
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

// use ReadBitsLE32 only when bits <= 32
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
