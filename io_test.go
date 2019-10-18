package fmp4parser

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func readBuf() ([]byte, error) {
	retBuf := make([]byte, 10240, 10240)
	fp, err := os.Open("D:\\aaa.mp4")
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	_, err = fp.Read(retBuf)
	if err != nil {
		err = errors.New("failed to read data")
		return nil, err
	}
	return retBuf, nil
}

func TestBufHandler_ReadInt(t *testing.T) {
	buf, err := readBuf()
	if err != nil {
		t.Error("failed on readBuf")
	}
	testHandler := NewBufHandler(buf)
	nRet := testHandler.ReadInt()
	// aaa.mp4's first 4-bytes is 0X (00 00 00 18) = 24 (Dec)
	if nRet != 24 {
		t.Errorf("readInt return %d want 24", nRet)
	}
}

func TestBufHandler_Shrink(t *testing.T) {
	buf, err := readBuf()
	if err != nil {
		t.Error("failed on readBuf")
	}
	testHandler := NewBufHandler(buf)

	fmt.Println(testHandler.valid, "   ", testHandler.index)
	nRet := testHandler.ReadInt()
	_ = testHandler.ReadInt()
	_, _ = testHandler.Move(int64(nRet) - 8)
	testHandler.Shrink()
	fmt.Println(testHandler.valid, "   ", testHandler.index)
	// should read 'moov' box's size
	nRet = testHandler.ReadInt()
	// aaa.mp4's 'moov' box's size  is 0X (00 00 0C 12)
	if nRet != 0xC12 {
		t.Errorf("readInt return %d want 0XC12", nRet)
	}
	testHandler.Append([]byte("1234"))
	fmt.Println(testHandler.valid, "   ", testHandler.index)
}

func TestBufHandler_FindBox(t *testing.T) {
	buf, err := readBuf()
	if err != nil {
		t.Error("failed on readBuf")
	}
	testHandler := NewBufHandler(buf)
	nRet, err := testHandler.FindBox(moovBox)
	if err != nil {
		t.Errorf("ReadInt failnRed")
	}
	if nRet != 0xC12 {
		t.Errorf("BufHandler_FindBox return %x, want 0XC12", nRet)
	}
}

func TestBufHandler_FindBoxInterval(t *testing.T) {
	buf, err := readBuf()
	if err != nil {
		t.Error("failed on readBuf")
	}
	testHandler := NewBufHandler(buf)
	nRet, err := testHandler.FindBox(moovBox)
	if err != nil {
		t.Errorf("FindBox failnRed")
	}
	if nRet != 0xC12 {
		t.Errorf("BufHandler_FindBox return %x, want 0XC12", nRet)
	}

	// find the 'trak' box
	nRet, err = testHandler.FindBoxInterval(trakBox, 0xC12-8)
	if err != nil {
		t.Errorf("FindBox failnRed")
	}
	if nRet != 0xB71 {
		t.Errorf("BufHandler_FindBoxInterval return %x, want 0XB71", nRet)
	}
}
