package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yangchenxing/cangshan/jaml"
)

func debug(s string) {
	os.Stderr.WriteString(s)
	os.Stderr.WriteString("\n")
}

func main() {
	jaml.EnableYaccErrorVerbose()
	if value, err := jaml.ParseStream(os.Stdin, debug, ""); err != nil {
		fmt.Fprintf(os.Stderr, "parse fail: %s\n", err.Error())
	} else if output, err := json.MarshalIndent(value, "", "    "); err != nil {
		fmt.Fprintf(os.Stderr, "marshal fail: %s\n", err.Error())
	} else {
		os.Stdout.WriteString(string(output))
		os.Stdout.WriteString("\n")
	}
}
