package dirtree

import (
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	root := t.TempDir()
	createDirStructure(t, root)

	got, err := Print(root, ModeAll)
	if err != nil {
		t.Fatal(err)
	}
	files := []string{
		"d 775 sym=0            crc=n/a      .",
		"d 775 sym=0            crc=n/a      A",
		"d 775 sym=0            crc=n/a      A/B",
		"? 777 sym=1            crc=n/a      A/B/symdirA",
		"f 775 sym=0 13b        crc=0451ac5e A/file1",
		"? 777 sym=1            crc=n/a      A/symfile1",
		"",
	}
	want := strings.Join(files, "\n")
	if got != want {
		t.Errorf("got:\n%v\nwant:\n%s", got, want)
	}
}
