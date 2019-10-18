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
func (p *BufHandler) ReadInt() uint32 {
	byteInt := make([]byte, 4, 4)
	anchor, _ := p.reader.Seek(0, io.SeekCurrent)
	_, err := io.ReadAtLeast(&p.reader, byteInt, 4)
	if err != nil {
		_, _ = p.reader.Seek(anchor, io.SeekStart)
		return 0xFFFF
	}
	p.index += 4
	return uint32(byteInt[0])<<24 | uint32(byteInt[1])<<16 | uint32(byteInt[2])<<8 | uint32(byteInt[3])
}

// ReadInt read 1 byte from buff slice and return the bitwise-integer if no error
func (p *BufHandler) ReadByte() int{
	byteInt := make([]byte, 1)
	anchor, _ := p.reader.Seek(0, io.SeekCurrent)
	_, err := io.ReadAtLeast(&p.reader, byteInt, 1)
	if err != nil {
		_, _ = p.reader.Seek(anchor, io.SeekStart)
		return 0xF
	}
	p.index += 4
	return int(byteInt[0])
}

// ReadInt read 2 byte from buff slice and return the bitwise-integer if no error
func (p *BufHandler) ReadShort() int {
	byteInt := make([]byte, 2)
	anchor, _ := p.reader.Seek(0, io.SeekCurrent)
	_, err := io.ReadAtLeast(&p.reader, byteInt, 2)
	if err != nil {
		_, _ = p.reader.Seek(anchor, io.SeekStart)
		return 0xFF
	}
	p.index += 4
	return int(byteInt[0])<<8 | int(byteInt[1])
}

// ReadInt read 8 byte from buff slice and return the bitwise-integer if no error
func (p *BufHandler) ReadLong() uint64 {
	byteInt := make([]byte, 8)
	anchor, _ := p.reader.Seek(0, io.SeekCurrent)
	_, err := io.ReadAtLeast(&p.reader, byteInt, 8)
	if err != nil {
		_, _ = p.reader.Seek(anchor, io.SeekStart)
		return 0XFFFFFFFF
	}
	p.index += 4
	high := uint32(byteInt[0])<<24 | uint32(byteInt[1])<<16 | uint32(byteInt[2])<<8 | uint32(byteInt[3])
	low := uint32(byteInt[4])<<24 | uint32(byteInt[5])<<16 | uint32(byteInt[6])<<8 | uint32(byteInt[7])
	return uint64(high)<<32 | uint64(low)
}


// ReadBytes read n bytes from buffer and return the n-bytes buffer
// and the size of n-bytes buffer if error is nil
func (p *BufHandler) ReadBytes(n int) ([]byte, int, error) {
	byteArr := make([]byte, n)
	anchor, _ := p.reader.Seek(0, io.SeekCurrent)
	nRet, err := io.ReadAtLeast(&p.reader, byteArr, n)
	if err != nil {
		_, _ = p.reader.Seek(anchor, io.SeekStart)
		return nil, 0, ErrNoEnoughData
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
		nSize := p.ReadInt()
		nType := p.ReadInt()
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
	return -1, err
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
		nSize := p.ReadInt()
		if interval < 4 {
			err = errors.New("FindBoxInterval: out of range")
			goto T
		}
		nType := p.ReadInt()
		if nType == boxtype {
			return int(nSize), nil
		} else {
			if interval < nSize-8 {
				err = errors.New("FindBoxInterval: out of range")
				goto T
			}
			interval -= nSize
			_, err = p.Move(int64(nSize - 8))
			if err != nil {
				goto T
			}
		}
	}
T:
	return -1, err
}

// Append is using for appending a slcie of byte to current p.b
func (p *BufHandler) Append(data []byte) {
	copy(p.b[p.valid:], data)
	p.valid += len(data)
}

// RemainSize return the size of un-read buffer
func (p *BufHandler) RemainSize() int {
	return p.valid - p.index
}

// GetCurrentBoxSize return the readed box's size( exclude box size and box type, total 8 byte)
// This function is intended to avoid adding arguments in some cases
func (p *BufHandler) GetCurrentBoxSize() int {
	_, _ = p.Move(-8)
	size := p.ReadInt()
	_, _ = p.Move(4)
	return int(size - 8)
}

// GetAbsPos return the distance between reader's pointer and the begin of buffer
func (p *BufHandler) GetAbsPos() int64 {
	index ,_ := p.reader.Seek(0,io.SeekCurrent)
	return index
}

func (p *BufHandler) SetPos(pos int64) error {
	if pos > int64(p.valid) {
		return ErrOutOfRange
	}
	index, err := p.reader.Seek(pos, io.SeekStart)
	if err != nil {
		return err
	}
	p.index = int(index)
	return nil
}