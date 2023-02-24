package main

import (
	"fmt"
	"io"
)

/*
Notice:
	The mp4Buffer in this file is modified from src/bytes/mp4Buffer.go which locates in Golang SDK.
	It only serves fmp4parser and perhaps is not universal for other project.
	If you need a mp4Buffer, please refer to the standard library(src/bytes/mp4Buffer.go).

	Q: Why should define the mp4Buffer?
	A: The bytes.Buffer is excellent. But fmp4parser needs some function which cannot be provided by bytes.Buffer,
		such as reading n bytes from io.Reader and moving back/ahead (i.e. Move) the reading pointer. These functions would
		provide convenience for operating mp4 structure buffer.
*/

// A mp4Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for buffer is an empty buffer ready to use.
// Non-Thread-Safe
type mp4Buffer struct {
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
}

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 128

const maxInt = int(^uint(0) >> 1)

// newMp4Buffer creates and initializes a new mp4Buffer using buf as its
// initial contents. The new mp4Buffer takes ownership of buf, and the
// caller should not use buf after this call. newMp4Buffer is intended to
// prepare a mp4Buffer to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(mp4Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a mp4Buffer.
func newMp4Buffer(buf []byte) *mp4Buffer { return &mp4Buffer{buf: buf} }

// ReadBytesFromAtLeast read nb bytes at least from readSeeker until EOF and appends it to the buffer,
// growing the buffer as needed. The return value n is the number of bytes read. The n should be equal with nb
// Any error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *mp4Buffer) ReadBytesFromAtLeast(r io.Reader, nb int) (n int, e error) {
	if nb <= 0 {
		return 0, nil
	}
	s := make([]byte, nb)
	n, e = r.Read(s)
	if n <= 0 || e != nil {
		return 0, fmt.Errorf("%w fmp4parser can not read bytes form outside Reader", e)
	}
	_, _ = b.Write(s)
	if n < nb {
		return 0, ErrNoEnoughData
	}
	return n, nil
}

// Move n bytes in the range of b.buf[:len(b.buf)] from current position.
// If the reading pointer not locate in the range, it will return false.
func (b *mp4Buffer) Move(n int) bool {
	if b.off+n < 0 || n > b.Len() {
		return false
	}
	b.off += n
	return true
}

// Peek reads the next len(p) bytes from the buffer.
// The return value n is the number of bytes read should be equal len(p). If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// If n != len(b), err is ErrNoEnoughData
func (b *mp4Buffer) Peek(p []byte) (n int, err error) {
	n, err = b.Read(p)
	if err != nil {
		return 0, err
	}
	b.off -= n
	return n, nil
}

// Read reads the next len(p) bytes from the mp4Buffer if there is sufficient data to read.
// The return value n is the number of bytes read and equal with len(p).
// If no data to read, return io.EOF.
// If no enough data to read, return ErrNoEnoughData
func (b *mp4Buffer) Read(p []byte) (n int, err error) {
	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	if b.Len() < len(p) {
		return 0, ErrNoEnoughData
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

/// -------------------- functions blow are from bytes package of Go SDK --------------------

// Bytes returns a slice of length b.Len() holding the unread portion of the mp4Buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like Read, Write, Reset, or Truncate).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (b *mp4Buffer) Bytes() []byte { return b.buf[b.off:] }

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the strings.Builder type.
func (b *mp4Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *mp4Buffer) Len() int { return len(b.buf) - b.off }

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (b *mp4Buffer) Cap() int { return cap(b.buf) }

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as Truncate(0).
func (b *mp4Buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

// empty reports whether the unread portion of the buffer is empty.
func (b *mp4Buffer) empty() bool { return len(b.buf) <= b.off }

// Write appends the contents of p to the buffer, growing the buffer when
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *mp4Buffer) Write(p []byte) (n int, err error) {
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// mp4Buffer without another allocation.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *mp4Buffer) Grow(n uint) {
	m := b.grow(int(n))
	b.buf = b.buf[:m]
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *mp4Buffer) grow(n int) int {
	m := b.Len()
	// If mp4Buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		// Not enough space anywhere, we need to allocate.
		buf := makeSlice(2*c + n)
		copy(buf, b.buf[b.off:])
		b.buf = buf
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *mp4Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error io.EOF.
func (b *mp4Buffer) ReadByte() (byte, error) {
	if b.empty() {
		// buffer is empty, reset to recover space.
		b.Reset()
		return 0, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	return c, nil
}
