package dirtree

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

func Write(w io.Writer, root string, mode PrintMode) error {
	bufw := bufio.NewWriter(w)

	return filepath.WalkDir(root, func(fullpath string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		line, err := mode.format(root, fullpath, dirent)
		if err != nil {
			return fmt.Errorf("WriteDirTree: can't format %s: %s", fullpath, err)
		}
		bufw.WriteString(line)
		bufw.WriteByte('\n')
		if err := bufw.Flush(); err != nil {
			return fmt.Errorf("WriteDirTree: can't write: %s", err)
		}
		return nil
	})
}

func Print(root string, mode PrintMode) (string, error) {
	var sb strings.Builder
	if err := Write(&sb, root, mode); err != nil {
		return "", err
	}
	return sb.String(), nil
}
