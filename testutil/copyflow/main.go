package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	remoteFilePattern = regexp.MustCompile("^(?P<username>[0-9])$")
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s remote_file url_prefix", filepath.Base(os.Args[0]))
		os.Exit(-1)
	}
}
