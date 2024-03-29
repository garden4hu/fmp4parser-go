package main

import (
	"fmt"
	"os"
	"testing"
)

func TestNewFmp4Parser(t *testing.T) {
	// Open file
	var moofUrl string = "C:/Users/peili/Desktop/js_hevc.mp4"
	fmt.Println("hu test begin")
	fileOP, err := os.Open(moofUrl)
	if err != nil {
		t.Fatal("failed to open source file.")
		return
	}

	defer fileOP.Close()
	fileStat, _ := fileOP.Stat()
	fmt.Println("xxxxx fileSize = ", fileStat.Size())
	// readSeeker := bufio.NewReader(fileOP) // doesn't need to use bufio
	// test
	var demuxer = NewFmp4Parser(fileOP) // logD.Print(test)
	// test_end
	_ = demuxer.Parse()

}
