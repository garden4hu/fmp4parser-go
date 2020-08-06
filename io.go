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
	br         *bytes.Reader // br for the cached data
}

func NewDeMuxReader(r io.ReadSeeker) *deMuxReader {
	de := deMuxReader{readSeeker: r}
	de.br = bytes.NewReader(de.buf)
	return &de
}

// ReadUnsignedByte read 1 byte and return  an uint8
func (p *deMuxReader) ReadUnsignedByte() uint8 {
	c, _ := p.br.ReadByte()
	return c
}

// Read2 read 2 byte from buff slice and return the bitwise-integer if no error
func (p *deMuxReader) Read2() uint16 {
	byteHeight := p.ReadUnsignedByte()
	byteLow := p.ReadUnsignedByte()
	return uint16(byteHeight)<<8 | uint16(byteLow)
}

// Read4 read 4 byte from buff slice and return the bitwise-integer if no error
func (p *deMuxReader) Read4() uint32 {
	bytesInt := make([]byte, 4, 4)
	// _, errLog := io.ReadAtLeast(&p.br, bytesInt, 4)
	_, _ = p.br.Read(bytesInt)
	return uint32(bytesInt[0])<<24 | uint32(bytesInt[1])<<16 | uint32(bytesInt[2])<<8 | uint32(bytesInt[3])
}

// Read8 read 8 byte from buff slice and return the bitwise-integer if no error
func (p *deMuxReader) Read8() uint64 {
	byteLong := make([]byte, 8, 8)
	_, _ = p.br.Read(byteLong)
	high := uint32(byteLong[0])<<24 | uint32(byteLong[1])<<16 | uint32(byteLong[2])<<8 | uint32(byteLong[3])
	low := uint32(byteLong[4])<<24 | uint32(byteLong[5])<<16 | uint32(byteLong[6])<<8 | uint32(byteLong[7])
	return uint64(high)<<32 | uint64(low)
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

// Shrink drop the used Data of buffer [remove form 0 to p.index]
func (p *deMuxReader) Shrink() {
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
		a, _ := p.GetAtomHeader()
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

// FindBoxInterval return the size of the box in the specific interval if error is nil
// Usually use this function to find the non-top Level box type, such as 'mvhd', 'esds' etc.
// Searching the top Level box type is also support.
func (p *deMuxReader) FindAtomWithinScope(boxType uint32, scope int64) (int64, error) {
	if scope < 8 {
		return -1, ErrAtomNotFound
	}
	leftSize := scope
	startPosition := p.Position()
	for leftSize > 8 {
		a, _ := p.GetAtomHeader()
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

// GetAtomHeader return the atom's type and size if no error encountered
// the atom.size doesn't contain the header size of the atom
func (p *deMuxReader) GetAtomHeader() (*atom, error) {
	var err error = nil
	if p.br.Len() <= 0 {
		err = p.Append(8)
	}
	var a atom
	a.atomSize = int64(p.Read4()) - 8
	a.atomType = p.Read4()
	if a.atomSize == 1 { // full box
		err = p.Append(8)
		a.atomSize = int64(p.Read8()) - 16
		a.atomHeaderSize = 16
	} else {
		a.atomHeaderSize = 8
	}
	return &a, err
}

// GetAtom return the next entire-atom's buffer
func (p *deMuxReader) GetNextAtomData() error {
	if p.Len() < 8 {
		if err := p.Append(8); err != nil {
			return err
		}
	}
	currentPosition := p.Position()
	a, _ := p.GetAtomHeader()
	err := p.CheckEnoughAtomData(a.atomSize)
	if err != nil {
		logE.Printf("failed to append atom:%s's body data", a.Type())
		_ = p.MoveTo(currentPosition)
		return err
	}
	_ = p.MoveTo(currentPosition)
	return err
}

// ReadVersionFlags return the Version of the box and the flags of the box
func (p *deMuxReader) ReadVersionFlags() (uint8, uint32) {
	n := p.Read4()
	return uint8(n >> 24 & 0xFF), n & 0x00FFFFFF
}

// CheckEnoughAtomData try to make sure the left data in buffer is enough.
// If not, deMuxReader will get more data from io. If there is error encountered,
// it will return error.
func (p *deMuxReader) CheckEnoughAtomData(atomSize int64) error {
	if atomSize > p.Len() {
		// try to read more buff( a.atomSize - p.Len() ) from io
		if p.Append(atomSize-p.Len()) != nil {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}

// Position return the distance between br's pointer and the begin of buffer
func (p *deMuxReader) Position() int64 {
	index, _ := p.br.Seek(0, io.SeekCurrent)
	return index
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
	p.Shrink() // clear cached data
	p.Reset()
}

// Skip try to skip n bytes un-read buffer
func (p *deMuxReader) Skip(offset int64) error {
	if offset < 0 {
		return ErrInvalidParam
	}
	currentPosition, _ := p.readSeeker.Seek(0, io.SeekCurrent)

	if p.Len() <= offset {
		_, err := p.readSeeker.Seek(offset-p.Len(), io.SeekCurrent)
		if err != nil {
			_, _ = p.readSeeker.Seek(currentPosition, io.SeekStart)
			return io.ErrUnexpectedEOF
		}
		p.ClearData()
	} else {
		p.Move(offset)
		p.Shrink()
	}
	return nil
}

// remove the data in the buffer
func (p *deMuxReader) ClearData() {
	p.Move(p.Len())
	p.Shrink()
}

// bitReader wraps an io.Reader and provides the ability to read values,
// bit-by-bit, from it. Its Read* methods don't return the usual error
// because the error handling was verbose. Instead, any error is kept and can
// be checked afterwards.
// Copy and modify from https://golang.org/src/compress/bzip2/bit_reader.go
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
