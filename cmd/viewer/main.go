package main

import (
	"fmt"
	"os"
)

func main() {
	folder := ""
	if len(os.Args) > 1 {
		folder = os.Args[1]
	}
	if folder == "" {
		fmt.Fprintln(os.Stderr, "usage: viewer <snapshot-folder>")
		fmt.Fprintln(os.Stderr, "example: viewer ./pkg/sketch/testdata")
	}
	LaunchViewer(folder)
}
