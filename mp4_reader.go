package fmp4parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// mp4Reader handles all things related to the buffer.
// It receives an io.Reader from fmp4parser API and reads un-parsed
// data from the latter.
type mp4Reader struct {
	r io.Reader
	b []byte // processing bytes
	a *atom  // processing atom
}

func newMp4Reader(i io.Reader) *mp4Reader {
	return &mp4Reader{
		r: i,
	}
}

func (p *mp4Reader) ReadAtomHeader() (a *atom, err error) {
	readInt := func(b []byte) uint32 {
		return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	}
	if p.a != nil {
		// if current atom's size is equal to len(p.b), it means read the next atom
		if p.a.headerSize == uint32(len(p.b)) {
			p.a = nil
			p.b = p.b[:0]
		}
	}
	if len(p.b) < 8 {
		header := make([]byte, 8-len(p.b))
		n, e := p.r.Read(header)
		if e != nil {
			return nil, e
		}
		p.b = append(p.b, header[:n]...)
		if n != len(header) {
			return nil, errors.New("cannot read more data from interface")
		}
	}

	a = new(atom)
	a.bodySize = int64(readInt(p.b[:4]))
	a.atomType = readInt(p.b[4:8])
	if a.bodySize == 1 {
		if len(p.b) < 16 {

			extHeader := make([]byte, 16-len(p.b))
			n, e := p.r.Read(extHeader)
			if e != nil {
				return nil, e
			}
			p.b = append(p.b, extHeader[:n]...)
			if n != len(extHeader) {
				return nil, errors.New("cannot read more data from interface")
			}
		}
		a.headerSize = 16
		a.bodySize = int64(readInt(p.b[8:12]))<<32 | int64(readInt(p.b[12:16])) - 16
	} else {
		a.bodySize -= 8
		a.headerSize = 8
	}
	p.a = a
	return a, nil
}

func (p *mp4Reader) ReadAtomBodyFull(body []byte) error {
	if len(p.b) == int(p.a.headerSize) { // doesn't has partial data
		n, e := p.r.Read(body)
		if e != nil {
			return e
		}
		if n != len(body) {
			p.b = append(p.b, body[:n]...)
			return errors.New("cannot read more data from interface")
		} else {
			// read OK, p.b wouldn't keep partial data of the atom
			p.b = p.b[:0]
			return nil
		}
	} else if len(p.b) > int(p.a.headerSize) {
		// it means there are some partial data
		need2Read := len(body) - len(p.b) + int(p.a.headerSize)
		if need2Read > 0 {
			left := make([]byte, need2Read)
			n, e := p.r.Read(left)
			if e != nil {
				return e
			}
			if n != len(left) {
				p.b = append(p.b, left[:n]...)
				return errors.New("cannot read more data from interface")
			}
			copy(body, p.b[p.a.headerSize:])
			copy(body[:len(p.b)-int(p.a.headerSize)], left)
		} else {
			copy(body, p.b[int(p.a.headerSize):len(body)+int(p.a.headerSize)])
		}
		p.b = p.b[:0]
		p.a = nil
		return nil
	} else {
		return errors.New("should get the header data firstly")
	}
}

// GetAtomReader return an atomReader if no error encountered.
// GetAtomReader will call ReadAtomBodyFull. So the error maybe isn't nil
func (p *mp4Reader) GetAtomReader(a *atom) (*atomReader, error) {
	body := make([]byte, a.bodySize)
	if e := p.ReadAtomBodyFull(body); e != nil {
		return nil, e
	}
	return newAtomReader(body, a), nil
}

func (p *mp4Reader) ReadAtomBody(body []byte) (n int, err error) {
	return p.r.Read(body)
}

// use it to read sample in "mdat"
func (p *mp4Reader) Read(b []byte) (n int, err error) {
	return p.r.Read(b)
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
