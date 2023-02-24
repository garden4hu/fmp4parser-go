package main

import (
	"fmt"
	"os"
)

var moofUrl string = "C:/Users/peili/Desktop/js_hevc.mp4"

func main() {
	fmt.Println("hu test begin")
	fileOP, err := os.Open(moofUrl)
	if err != nil {
		println("failed to open source file.")
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
