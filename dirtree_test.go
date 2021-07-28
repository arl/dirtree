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

	lines := []string{
		"d            crc=n/a      .",
		"d            crc=n/a      A",
		"d            crc=n/a      A/B",
		"?            crc=n/a      A/B/symdirA",
		"f 13b        crc=0451ac5e A/file1",
		"?            crc=n/a      A/symfile1",
	}

	got = strings.TrimSpace(got)
	if want := strings.Join(lines, "\n"); got != want {
		t.Errorf("got:\n%v\nwant:\n%s", got, want)
	}
}
