package dirtree_test

import (
	"fmt"
	"testing"

	"github.com/arl/dirtree"
)

func TestPrintDirTree(t *testing.T) {
	dt, err := dirtree.PrintDirTree("../statsviz/_example", dirtree.ModeAll)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(dt)
}
