package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, " %s live_stream_url\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	streamInfoChan := make(chan *streamInfo)

	go demuxing(flag.Arg(0), streamInfoChan)

	drawUI(streamInfoChan)
}
