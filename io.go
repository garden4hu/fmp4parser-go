package fmp4parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
)

type bufferHandler struct {
	buf    []byte       // buffer
	reader bytes.Reader // read buffer
	mux    sync.Mutex
}

// newBufHandler return the object the operate the buffer
func newBufHandler(b []byte) *bufferHandler {
	bufferT := new(bufferHandler)
	bufferT.buf = make([]byte, len(b))
	copy(bufferT.buf, b)
	bufferT.reader = *(bytes.NewReader(bufferT.buf))
	return bufferT
}

// Read4 read 4 byte from buff slice and return the bitwise-integer if no error
func (p *bufferHandler) Read4() uint32 {
	byteInt := make([]byte, 4, 4)
	//_, err := io.ReadAtLeast(&p.reader, byteInt, 4)
	_, _ = p.reader.Read(byteInt)
	return uint32(byteInt[0])<<24 | uint32(byteInt[1])<<16 | uint32(byteInt[2])<<8 | uint32(byteInt[3])
}

// Read4 read 1 byte from buff slice and return the bitwise-integer if no error
func (p *bufferHandler) ReadByte() int {
	cbyte, _ := p.reader.ReadByte()
	return int(cbyte)
}

// Read2 read 2 byte from buff slice and return the bitwise-integer if no error
func (p *bufferHandler) Read2() int {
	byteShort := make([]byte, 2, 2)
	_, _ = p.reader.Read(byteShort)
	return int(byteShort[0])<<8 | int(byteShort[1])
}

// Read8 read 8 byte from buff slice and return the bitwise-integer if no error
func (p *bufferHandler) Read8() uint64 {
	byteLong := make([]byte, 8, 8)
	_, _ = p.reader.Read(byteLong)
	high := uint32(byteLong[0])<<24 | uint32(byteLong[1])<<16 | uint32(byteLong[2])<<8 | uint32(byteLong[3])
	low := uint32(byteLong[4])<<24 | uint32(byteLong[5])<<16 | uint32(byteLong[6])<<8 | uint32(byteLong[7])
	return uint64(high)<<32 | uint64(low)
}

// ReadBytes read n bytes from buffer and return the n-bytes buffer
// and the size of n-bytes buffer if error is nil
func (p *bufferHandler) ReadBytes(n int) ([]byte, int, error) {
	if n <= 0 {
		return nil, -1, errors.New("invalid param: n")
	}
	byteArr := make([]byte, n, n)
	nRet, _ := p.reader.Read(byteArr)
	return byteArr, nRet, nil
}

// Cut drop the used data of buffer [remove form 0 to p.index]
func (p *bufferHandler) Cut() {
	p.mux.Lock()
	currentIndex, _ := p.reader.Seek(0, io.SeekCurrent)
	copy(p.buf[0:], p.buf[currentIndex:])
	for k, n := int64(len(p.buf))-currentIndex, int64(len(p.buf)); k < n; k++ {
		p.buf[k] = 0
	}
	p.buf = p.buf[:int64(len(p.buf))-currentIndex]
	p.reader.Reset(p.buf)
	p.mux.Unlock()
}

// Reset reset the reader pointer to the begin of slice
func (p *bufferHandler) Reset() {
	p.mux.Lock()
	p.reader.Reset(p.buf)
	p.mux.Unlock()
}

// Move wrap the io.Seeker without error check
func (p *bufferHandler) Move(siz int64) {
	_, _ = p.reader.Seek(siz, io.SeekCurrent)
}

// Peek return the next n byte unread data without change the reader's index.
// if the condition is not met, return error
func (p *bufferHandler) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrOutOfRange
	}
	if n > p.reader.Len() {
		return nil, ErrRequestTooLarge
	}
	currentIndex, _ := p.reader.Seek(0, io.SeekCurrent)
	return p.buf[currentIndex : currentIndex+int64(n)], nil
}

// FindBox return the size of the specific type of iso box if error equal nil
// usually use this function to find the top level box type, such as 'ftyp', 'moov' etc.
func (p *bufferHandler) FindBox(boxtype uint32) (int, error) {
	err := errors.New("")
	for {
		nSize := p.Read4()
		nType := p.Read4()
		if nType == boxtype {
			return int(nSize), nil
		} else {
			_, err := p.reader.Seek(int64(nSize-8), io.SeekCurrent)
			if err != nil {
				goto T
			}
		}
	}
T:
	return -1, err
}

// FindBoxInterval return the size of the box in the specific interval if error is nil
// Usually use this function to find the non-top level box type, such as 'mvhd', 'esds' etc.
// Searching the top level box type is also support.
func (p *bufferHandler) FindBoxInterval(boxtype uint32, interval uint32) (int, error) {
	err := errors.New("")
	for {
		if interval < 4 {
			err = errors.New("FindBoxInterval: out of range")
			goto T
		}
		nSize := p.Read4()
		if interval < 4 {
			err = errors.New("FindBoxInterval: out of range")
			goto T
		}
		nType := p.Read4()
		if nType == boxtype {
			return int(nSize), nil
		} else {
			if interval < nSize-8 {
				err = errors.New("FindBoxInterval: out of range")
				goto T
			}
			interval -= nSize
			_, err := p.reader.Seek(int64(nSize-8), io.SeekCurrent)
			if err != nil {
				goto T
			}
		}
	}
T:
	return -1, err
}

// Append is using for appending a slcie of byte to current p.buf
func (p *bufferHandler) Append(data []byte) {
	p.mux.Lock()
	copy(p.buf[len(p.buf):], data)
	p.reader.Reset(p.buf)
	p.mux.Unlock()
}

// Remainder return the size of un-read buffer
func (p *bufferHandler) Remainder() int {
	return p.reader.Len()
}

// GetCurrentBoxHeaderInfo return the readed box's size( exclude box size and box type, total 8 byte)
// This function is intended to avoid adding arguments in some cases
func (p *bufferHandler) GetCurrentBoxHeaderInfo() (int64, uint32) {
	p.Move(-8)
	size := p.Read4()
	p.Move(4)
	return int64(size - 8), 0
}

// Position return the distance between reader's pointer and the begin of buffer
func (p *bufferHandler) Position() int {
	index, _ := p.reader.Seek(0, io.SeekCurrent)
	return int(index)
}

func (p *bufferHandler) MoveTo(pos int) error {
	if pos > len(p.buf) {
		return ErrOutOfRange
	}
	_, err := p.reader.Seek(int64(pos), io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}

// bitReader wraps an io.Reader and provides the ability to read values,
// bit-by-bit, from it. Its Read* methods don't return the usual error
// because the error handling was verbose. Instead, any error is kept and can
// be checked afterwards.
// Modified by https://golang.org/src/compress/bzip2/bit_reader.go
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

func newBitReaderFromString(src string) bitReader {

	return newBitReader(strings.NewReader(src))
}

func (br *bitReader) ReadBits64(bits uint) (n uint64) {
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

func (br *bitReader) ReadBits(bits uint) (n int) {
	n64 := br.ReadBits64(bits)
	return int(n64)
}

func (br *bitReader) ReadBit() bool {
	n := br.ReadBits(1)
	return n != 0
}

func (br *bitReader) Err() error {
	return br.err
}
