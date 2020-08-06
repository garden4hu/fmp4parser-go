package fmp4parser

import (
	"fmt"
	"os"
	"testing"
)

var moofUrl string = "D:\\moov\\sintel_3min_rotate_audio.mp4"

func TestNewFmp4Parser(t *testing.T) {
	// Open file
	fileOP, err := os.Open(moofUrl)
	if err != nil {
		t.Fatal("failed to open source file.")
		return
	}
	fileStat, _ := fileOP.Stat()
	fmt.Println("xxxxx fileSize = ", fileStat.Size())
	// r := bufio.NewReader(fileOP) // doesn't need to use bufio
	// test
	demux := NewFmp4Parser(fileOP)
	// logD.Print(test)
	// test_end
	_ = demux.Parse()
}
