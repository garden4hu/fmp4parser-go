package fmp4parser

import (
	"os"
	"testing"
)

var targetNumber int = 1024

func TestBuffer_Read(t *testing.T) {
	sl := make([]byte, 0, 4)
	b := newMp4Buffer(sl)

	filename := "test_file_large"
	fp, e := os.Open(filename)
	if e != nil {
		t.Errorf("failed to open test file: %s\n", filename)
		return
	}
	defer func(fp *os.File) { _ = fp.Close() }(fp)
	n, e := b.ReadBytesFromAtLeast(fp, targetNumber)
	if e != nil {
		t.Fatal("ReadBytesFrom return failed")
	}
	if n != targetNumber {
		t.Fatalf("ReadBytesFrom Actually input : %d\n", n)
	}

	sR := make([]byte, targetNumber/2)
	nR, e := b.Read(sR)
	if e != nil {
		t.Fatal("Read failed", e)
	}
	if b.Len() != (targetNumber - nR) {
		t.Fatal("internal error: after read")
	}
}
