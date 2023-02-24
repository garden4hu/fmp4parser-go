package main

import (
	"bytes"
	"io"
)

type atomReader struct {
	b []byte
	r *bytes.Reader
	a *atom
}

func newAtomReader(b []byte, a *atom) *atomReader {
	ar := new(atomReader)
	ar.b = b
	ar.r = bytes.NewReader(ar.b)
	ar.a = a
	return ar
}

// ReadUnsignedByte read 1 byte and return uint8
func (p *atomReader) ReadUnsignedByte() uint8 {
	c, _ := p.r.ReadByte()
	return c
}

// ReadSignedByte read 1 byte and return int8
func (p *atomReader) ReadSignedByte() int8 {
	c, _ := p.r.ReadByte()
	return int8(c)
}

// Read2 read 2 byte from mp4Buffer slice and return the bitwise-integer if no error
func (p *atomReader) Read2() uint16 {
	h := p.ReadUnsignedByte()
	l := p.ReadUnsignedByte()
	return uint16(h)<<8 | uint16(l)
}

// Read2S read 2 byte from mp4Buffer slice and return a signed integer(16bits)if no error
func (p *atomReader) Read2S() int16 {
	h := p.ReadSignedByte()
	l := p.ReadUnsignedByte()
	return int16(h)<<8 | int16(l)
}

// Read4 read 4 bytes from mp4Buffer slice and return an unsigned integer(32bits) if no error
func (p *atomReader) Read4() uint32 {
	h := p.Read2()
	l := p.Read2()
	return uint32(h)<<16 | uint32(l)
}

// Read4S read 4 bytes from mp4Buffer slice and return a signed integer(32bits) if no error
func (p *atomReader) Read4S() int32 {
	h := p.Read2S()
	l := p.Read2()
	return int32(h)<<16 | int32(l)
}

// Read8 read 8 bytes from mp4Buffer slice and return an unsigned integer(64bits) if no error
func (p *atomReader) Read8() uint64 {
	h := p.Read4()
	l := p.Read4()
	return uint64(h)<<32 | uint64(l)
}

// Read8S read 8 bytes from mp4Buffer slice and return a signed integer(64bits) if no error
func (p *atomReader) Read8S() int64 {
	h := p.Read4S()
	l := p.Read4()
	return int64(h)<<32 | int64(l)
}

// ReadBytes read n bytes if error is nil
func (p *atomReader) ReadBytes(b []byte) (int, error) {
	return p.r.Read(b)
}

func (p *atomReader) Peek(b []byte) error {
	cur, _ := p.r.Seek(0, io.SeekCurrent)
	n, err := p.r.Read(b)
	_, _ = p.r.Seek(cur, io.SeekStart)
	if err != nil || n != len(b) {
		return ErrNoEnoughData
	}
	return nil
}

// ReadAtomHeader parseConfig the atom from the buffer
func (p *atomReader) ReadAtomHeader() *atom {
	a := new(atom)
	fullAtomSize := int64(p.Read4())
	a.atomType = p.Read4()
	if fullAtomSize == 1 { // full box
		fullAtomSize = int64(p.Read8())
		a.headerSize = 16
	} else {
		a.headerSize = 8
	}
	a.bodySize = fullAtomSize - int64(a.headerSize)
	return a
}

// ReadVersionFlags return the Version of the box and the flags of the box
func (p *atomReader) ReadVersionFlags() (uint8, uint32) {
	n := p.Read4()
	return uint8(n >> 24 & 0xFF), n & 0x00FFFFFF
}

func (p *atomReader) Len() int {
	return p.r.Len()
}

func (p *atomReader) Size() int {
	return int(p.r.Size())
}
func (p *atomReader) Move(n int) error {
	current, _ := p.r.Seek(0, io.SeekCurrent)
	if _, err := p.r.Seek(int64(n), io.SeekCurrent); err != nil {
		_, _ = p.r.Seek(current, io.SeekStart)
		return err
	}
	return nil
}

// FindSubAtom will find the atomType. If no error encountered, a atomReader of the
// atom will return and the reader pointer will pointer to the flowing atom of the atomType.
// If atomType not found, ErrAtomNotFound will be returned. The internal state will keep unchanging.
// If ErrAtomSizeInvalid returned, it means that an unacceptable error has occurred
// FindSubAtom only return a sub-atomReader and will not change the internal state.
func (p *atomReader) FindSubAtom(atomType uint32) (ar *atomReader, err error) {
	if p.r.Len() < 8 {
		return nil, ErrAtomNotFound
	}
	cur, _ := p.r.Seek(0, io.SeekCurrent)
	ar = nil
	err = nil
	for p.r.Len() > 8 {
		start, _ := p.r.Seek(0, io.SeekCurrent)
		a := p.ReadAtomHeader()
		if a.bodySize <= int64(p.r.Len()) {
			if a.atomType == atomType {
				ar = newAtomReader(p.b[(start+int64(a.headerSize)):start+a.Size()], a)
				break
			} else {
				_, _ = p.r.Seek(a.bodySize, io.SeekCurrent)
			}
		} else {
			err = ErrInvalidAtomSize
			break
		}
	}
	//  if not found, rest the reader
	_, _ = p.r.Seek(cur, io.SeekStart)
	return ar, err
}

func (p *atomReader) GetNextAtom() (*atomReader, error) {
	if p.r.Len() == 0 {
		return nil, ErrNoMoreAtom
	}
	if p.r.Len() < 8 {
		return nil, ErrNoEnoughData
	}
	start, _ := p.r.Seek(0, io.SeekCurrent)
	a := p.ReadAtomHeader()
	if a.bodySize <= int64(p.r.Len()) {
		_, _ = p.r.Seek(a.bodySize, io.SeekCurrent)
		ar := newAtomReader(p.b[(start+int64(a.headerSize)):start+a.Size()], a)
		ar.a = a
		return ar, nil
	} else {
		_, _ = p.r.Seek(start, io.SeekStart)
		return nil, ErrNoEnoughData
	}
}
