package fmp4parser

import (
	"bytes"
	"errors"
	"io"
)

type BufHandler struct {
	b      []byte       // buffer
	index  int          // current index
	reader bytes.Reader // read buffer
	valid  int          // valid data size
}

func NewBufHandler(b []byte) *BufHandler {
	return &BufHandler{b: b, index: 0, reader: *(bytes.NewReader(b)), valid: len(b)}
}

// ReadInt read 4 byte from buff slice and return the bitwise-integer if no error
func (p *BufHandler) ReadInt() (uint32, error) {
	byteInt := make([]byte, 4, 4)
	_, err := io.ReadAtLeast(&p.reader, byteInt, 4)
	if err != nil {
		return 0, ErrEof
	}
	p.index += 4
	return uint32(byteInt[0])<<24 | uint32(byteInt[1])<<16 | uint32(byteInt[2])<<8 | uint32(byteInt[3]), nil
}

// ReadBytes read n bytes from buffer and return the n-bytes buffer
// and the size of n-bytes buffer if error is nil
func (p *BufHandler) ReadBytes(n int) ([]byte, int, error) {
	byteArr := make([]byte, 0, n)
	nRet, err := io.ReadAtLeast(&p.reader, byteArr, n)
	if err != nil {
		if err == io.ErrShortBuffer {
			return nil, 0, io.ErrShortBuffer
		} else {
			return nil, -1, ErrUnexpectedEof
		}
	}
	p.index += nRet
	return byteArr, nRet, nil
}

// Shrink drop the used data of buffer [remove form 0 to p.index]
func (p *BufHandler) Shrink() {
	copy(p.b[0:], p.b[p.index:])
	for k, n := len(p.b)-p.index, len(p.b); k < n; k++ {
		p.b[k] = 0
	}
	p.b = p.b[:len(p.b)-p.index]
	p.valid -= p.index
	p.index = 0
	p.reader.Reset(p.b)
}

// ResetReader reset the reader pointer to the begin of slice
func (p *BufHandler) ResetReader() {
	p.reader.Reset(p.b)
}

// Move wrap the io.Seeker
func (p *BufHandler) Move(siz int64) (int64, error) {
	nRet, err := p.reader.Seek(siz, io.SeekCurrent)
	if err == nil {
		p.index = int(nRet)
		return nRet, nil
	} else {
		return 0, err
	}
}

// FindBox return the size of the specific type of iso box if error equal nil
// usually use this function to find the top level box type, such as 'ftyp', 'moov' etc.
func (p *BufHandler) FindBox(boxtype uint32) (int, error) {
	err := errors.New("")
	for {
		nSize, err := p.ReadInt()
		if err != nil {
			goto T
		}
		nType, err := p.ReadInt()
		if err != nil {
			goto T
		}
		if nType == boxtype {
			return int(nSize), nil
		} else {
			_, err = p.Move(int64(nSize - 8))
			if err != nil {
				goto T
			}
		}
	}
T:
	if err == io.ErrShortBuffer {
		return -1, io.ErrShortBuffer
	} else {
		return -1, err
	}
}

// FindBoxInterval return the boxtype's size in the specific interval if error is nil
// Usually use this function to find the non-top level box type, such as 'mvhd', 'esds' etc.
// Searching the top level box type is also support.
func (p *BufHandler) FindBoxInterval(boxtype uint32, interval uint32) (int, error) {
	err := errors.New("")
	for {
		if interval < 4 {
			err = errors.New("FindBoxInterval: out of range")
			goto T
		}
		nSize, err := p.ReadInt()
		if err != nil {
			goto T
		}
		interval -= 4

		if interval < 4 {
			err = errors.New("FindBoxInterval: out of range")
			goto T
		}
		nType, err := p.ReadInt()
		if err != nil {
			goto T
		}

		if nType == boxtype {
			return int(nSize), nil
		} else {
			interval -= 4
			if interval < nSize-8 {
				err = errors.New("FindBoxInterval: out of range")
				goto T
			}
			_, err = p.Move(int64(nSize - 8))
			if err != nil {
				goto T
			}
		}
	}
T:
	if err == io.ErrShortBuffer {
		return -1, io.ErrShortBuffer
	} else {
		return -1, err
	}
}

func (p *BufHandler) Append(data []byte) {
	copy(p.b[p.valid:], data)
	p.valid += len(data)
}
