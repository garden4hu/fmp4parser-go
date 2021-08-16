package fmp4parser

import (
	"errors"
	"sync/atomic"
)

// bufferCache is a simple queue for mp4Buffer used for fmp4parser

const SlotEntry = 1024
const SlotSize = 4096

// internal mp4Buffer size is 4M == SlotEntry*SlotSize bytes
// bufferCache.b is a consecutive bytes
// How it works: Slice b is divided into 1024 sub-slices, call it slot. Each slot
// has 4096 bytes memory. readingSlot and writingSlot respectively
//represent the slot where the reading pointer and the writing pointer are located.
// readingIndex and writingIndex respectively represent the distance from the corresponding pointer to the slot head.
// The most important thing: the reading pointer will never catch up with the writing pointer.
// You can think of this structure as a variant of circular mp4Buffer.
type bufferCache struct {
	b            [][]byte // SlotEntry slot x SlotSize bytes
	readingSlot  int      // the slot number of reading currently
	writingSlot  int      // the slot number of writing currently
	readingIndex int      // point to unread data of readingSlot
	writingIndex int      // point to the byte for writing  of writingIndex
	absPosition  int64    // the start position in origin file of slice b  : form writingSlot.writingIndex to writingSlot.(writingIndex-1)
	length       int32
}

func newBufferCache() *bufferCache {
	b := make([][]byte, SlotEntry)
	for i := range b {
		b[i] = make([]byte, SlotSize)
	}
	return &bufferCache{b: b}
}

// Len is alias of length
func (p *bufferCache) Len() int {
	return int(atomic.LoadInt32(&p.length))
}

// Read n bytes from bufferCache.
// Read implements the io.Reader interface.
// Notice: Non-Thread-Safety
func (p *bufferCache) Read(b []byte) (n int, e error) {
	currentLen := int(atomic.LoadInt32(&p.length))
	retrieveData := func() {
		leftToRead := n
		nRead := 0
		if leftToRead < SlotSize-p.readingIndex {
			nRead = copy(b, (p.b[p.readingSlot])[p.readingIndex:p.readingIndex+leftToRead])
			p.readingIndex += nRead
			p.readingIndex %= SlotSize
		} else {
			nRead = copy(b, (p.b[p.readingSlot])[p.readingIndex:])
			p.readingIndex += leftToRead
			p.readingIndex %= SlotSize
			p.readingSlot++
			p.readingSlot %= SlotEntry
			leftToRead -= nRead
			leftSlotToRead := leftToRead / SlotSize
			residual := leftToRead % SlotSize
			for i := 0; i < leftSlotToRead; i++ {
				nRead += copy(b[nRead:], (p.b[p.readingSlot])[:])
				p.readingSlot++
				p.readingSlot %= SlotEntry
			}
			if residual > 0 {
				copy(b[nRead:], (p.b[p.readingSlot])[:residual])
			}
			p.readingIndex += residual
			p.readingIndex %= SlotSize
		}
		atomic.AddInt32(&p.length, int32(-n))
	}
	n = 0
	e = nil
	if atomic.LoadInt32(&p.length) <= 0 {
		e = errors.New("no enough data to read")
	} else {
		if len(b) >= currentLen {
			n = currentLen
		} else {
			n = len(b)
		}
		retrieveData()
	}
	return n, e
}

// Write will attach len(b) bytes data to the tail of bufferCache.b
// As we set, the upper bound is 4M, and len(b) is far small with it.
// Write implements the io.Writer interface.
// Notice: Non-Thread-Safety
func (p *bufferCache) Write(b []byte) (n int, e error) {
	currentLen := int(atomic.LoadInt32(&p.length))
	appendData := func(b []byte) {
		currentWritingSlotLeft := SlotSize - p.writingIndex
		nWritten := 0
		if len(b) < currentWritingSlotLeft {
			nWritten = copy((p.b[p.writingSlot])[p.writingIndex:], b)
			p.writingIndex += nWritten
			p.writingIndex %= SlotSize
		} else {
			nWritten = copy((p.b[p.writingSlot])[p.writingIndex:], b)
			p.writingIndex += nWritten
			p.writingIndex %= SlotSize
			p.writingSlot++
			p.writingSlot %= SlotEntry
			leftToWrite := len(b) - nWritten
			leftSlotToWrite := leftToWrite / SlotSize
			residual := leftToWrite % SlotSize
			for i := 0; i < leftSlotToWrite; i++ {
				nWritten += copy(p.b[p.writingSlot], b[nWritten:])
				p.writingSlot++
				p.writingSlot %= SlotEntry
			}
			if residual > 0 {
				_ = copy(p.b[p.writingSlot], b[nWritten:])
			}
			p.writingIndex += residual
			p.writingIndex %= SlotSize
		}
		atomic.AddInt32(&p.length, int32(len(b)))
	}
	n = 0
	e = nil
	if SlotSize*SlotEntry-currentLen > 0 {
		if SlotSize*SlotEntry-currentLen <= len(b) {
			b2 := b[:SlotSize*SlotEntry-currentLen]
			appendData(b2)
			n = SlotSize*SlotEntry - currentLen
		} else {
			appendData(b)
			n = len(b)
		}
	} else {
		e = errors.New("no more space to write")
	}
	return n, e
}

// Reset will reset internal value
func (p *bufferCache) Reset() {
	p.readingIndex = 0
	p.readingSlot = 0
	p.writingIndex = 0
	p.writingSlot = 0
	atomic.StoreInt32(&p.length, 0)
}
