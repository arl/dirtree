package dirtree

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	mode     PrintMode
	showRoot bool
}

var defaultCfg = config{
	mode:     ModeAll,
	showRoot: true,
}

type Option interface {
	apply(*config) error
}

// The ExcludeRoot option hides the root directory from the list.
var ExcludeRoot Option = IncludeRoot(false)

// ExcludeRoot is the option controlling whether the root directory should be
// printed when listing its content.
type IncludeRoot bool

func (in IncludeRoot) apply(cfg *config) error {
	cfg.showRoot = bool(in)
	return nil
}



func Write(w io.Writer, root string, opts ...Option) error {
	cfg := defaultCfg
	for _, o := range opts {
		if err := o.apply(&cfg); err != nil {
			return fmt.Errorf("dirtree: configuration error: %v", err)
		}
	}

	seenRoot := false

	bufw := bufio.NewWriter(w)
	err := filepath.WalkDir(root, func(fullpath string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Exclude root
		if !seenRoot {
			seenRoot = true
			if !cfg.showRoot {
				return nil
			}
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

// Sprint calls Write and returns the list of files as a string.
func Sprint(root string, opts ...Option) (string, error) {
	var sb strings.Builder
	if err := Write(&sb, root, opts...); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Print is a wrapper around Write(os.Stdout, ...).
func Print(root string, opts ...Option) error {
	return Write(os.Stdout, root, opts...)
}
