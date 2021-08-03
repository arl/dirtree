package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arl/dirtree"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("[dirtree] ")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "dirtree recursively lists a directory content")
		fmt.Fprintln(os.Stderr, "usage: dirtree [DIR]")
		fmt.Fprintln(os.Stderr, "\tDIR defaults to current directory")
	}
	flag.Parse()

	dir := "."
	if flag.NArg() == 1 {
		dir = flag.Args()[0]
	}
	if flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	if err := dirtree.Write(os.Stdout, dir, dirtree.ModeAll); err != nil {
		log.Fatalf("error: %v", err)
	}
}
