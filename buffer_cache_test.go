package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"
)

func TestBufferCache_Read(t *testing.T) {
	filename := "test_file_large"
	fp, e := os.Open(filename)
	if e != nil {
		t.Errorf("failed to open test file: %s\n", filename)
		return
	}
	defer func(fp *os.File) { _ = fp.Close() }(fp)
	bc := newBufferCache()
	ch := make(chan bool)

	//pprof
	go func() {
		fmt.Println(http.ListenAndServe("localhost:8000", nil))
	}()
	consumed := make(chan int64)
	go consumer(ch, consumed, bc)
	total := int64(0)
	cancel := func(chan bool) { ch <- true }
	s := make([]byte, SlotSize)
	for {
		// time.Sleep(time.Microsecond * time.Duration(10))
		n, e := fp.Read(s)
		if e != nil && e == io.EOF {
			log.Println("producer Load file finish")
			goto L
		}
		ret := 0
		for i := 0; i < 3; i++ {
			nW, e := bc.Write(s[ret:n])
			if e != nil {
				t.Error(e)
				// time.Sleep(time.Microsecond * time.Duration(10))
				continue
			}
			// log.Printf("INFO : producer written %d bytes OK\n", nW)
			total += int64(nW)
			ret += nW
			if ret == n {
				break
			}
		}
	}
L:
	log.Printf("INFO : producer written %d bytes\n", total)
	cancel(ch)
	time.Sleep(time.Second * time.Duration(1)) // 2s is enough for reading 4M bytes
	consumedSize := <-consumed

	if total != consumedSize {
		t.Errorf("test failed. produce %d , consume %d\n", total, consumedSize)
	}
}

func consumer(cancel chan bool, ret chan int64, cache *bufferCache) {
	if cache == nil {
		return
	}
	total := int64(0)
	s := make([]byte, SlotSize)
	customerInternal := func() {
		if cache.Len() <= 0 {
			return
		}
		n, e := cache.Read(s)
		if e != nil {
			log.Printf("WARN: consumer faild to read data\n")
			return
		}
		// fmt.Printf("INFO: consumer  read %d bytes ok\n", n)
		total += int64(n)
	}

	for {
		// time.Sleep(time.Microsecond * time.Duration(10))
		select {
		case <-cancel:
			log.Printf("INFO: consumer read %d bytes\n", total)
			ret <- total
			return
		default:
			customerInternal()
		}
	}
}
