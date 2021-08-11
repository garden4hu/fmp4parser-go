package fmp4parser

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestBufferCache_Read(t *testing.T) {
	filename := string("D:/Test/test_large_file")
	fp, e := os.Open(filename)
	if e != nil {
		t.Errorf("failed to open test file: %s", filename)
		return
	}
	bc := newBufferCache()

	ch := make(chan bool)

	go customer(ch, bc)

	cancel := func(chan bool) { ch <- true }
	s := make([]byte, SlotSize)
	for {
		n, e := fp.Read(s)
		if e != nil {
			t.Error("producer read error from test file")
			cancel(ch)
			return
		}
		for i := 0; i < 3; i++ {
			nW, e := bc.Write(s[:n])
			if i == 3 {
				t.Error("producer failed to write for 3 times. Testing failed.")
				cancel(ch)
				goto L
			}
			if e != nil {
				t.Error(e)
				time.Sleep(time.Microsecond * time.Duration(10))
				continue
			}
			fmt.Printf("INFO : producer written %d bytes OK", nW)
		}
	}
L:
	return
}

func customer(cancel chan bool, cache *bufferCache) {
	if cache == nil {
		return
	}
	s := make([]byte, SlotSize)
	customerInternal := func() {
		if cache.Len() <= 0 {
			return
		}
		n, e := cache.Read(s)
		if e != nil {
			fmt.Printf("WARN: faild to read data")
			return
		}
		fmt.Printf("INFO: customer  read %d bytes ok", n)
	}
	for {
		select {
		case <-cancel:
			return
		default:
			customerInternal()
		}
	}
}
