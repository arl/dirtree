package dirtree

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type Option interface {
	apply(*config) error
}

type config struct {
	mode PrintMode
}

var defaultCfg = config{
	mode: ModeAll,
}

func Write(w io.Writer, root string, opts ...Option) error {
	cfg := defaultCfg
	for _, o := range opts {
		if err := o.apply(&cfg); err != nil {
			return fmt.Errorf("dirtree: configuration error: %v", err)
		}
	}

	bufw := bufio.NewWriter(w)
	err := filepath.WalkDir(root, func(fullpath string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		line, err := cfg.mode.format(root, fullpath, dirent)
		if err != nil {
			return fmt.Errorf("can't format %s: %s", fullpath, err)
		}
		bufw.WriteString(line)
		bufw.WriteByte('\n')
		return nil
	})
	if err != nil {
		return fmt.Errorf("dirtree: error walking directory: %v", err)
	}
	if err := bufw.Flush(); err != nil {
		return fmt.Errorf("dirtree: can't write output: %s", err)
	}
	return nil
}

func Print(root string, opts ...Option) (string, error) {
	var sb strings.Builder
	if err := Write(&sb, root, opts...); err != nil {
		return "", err
	}
	return sb.String(), nil
}
