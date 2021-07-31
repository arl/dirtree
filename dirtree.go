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

type writer struct {
	cfg      config
	seenRoot bool
	root     string
	bufw     *bufio.Writer
}

func newWriter(w io.Writer, root string, opts ...Option) (*writer, error) {
	cfg := defaultCfg
	for _, o := range opts {
		if err := o.apply(&cfg); err != nil {
			return nil, fmt.Errorf("dirtree: configuration error: %v", err)
		}
	}

	return &writer{
		bufw:     bufio.NewWriter(w),
		seenRoot: false,
		root:     root,
		cfg:      cfg,
	}, nil
}

func (w *writer) walk(fullpath string, dirent fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	// Exclude root
	if !w.seenRoot {
		w.seenRoot = true
		if !w.cfg.showRoot {
			return nil
		}
	}

	// Path conversion: relative to root and slash based
	rel, err := filepath.Rel(w.root, fullpath)
	if err != nil {
		return err
	}

	// Depth check
	if w.cfg.depth != 0 {
		if len(strings.Split(rel, string(os.PathSeparator))) > w.cfg.depth {
			if dirent.IsDir() {
				err = fs.SkipDir
			}
			return err
		}
	}

	rel = filepath.ToSlash(rel)

	// Ignore patterns
	if w.cfg.ignore != nil {
		for _, pattern := range w.cfg.ignore {
			if m, _ := filepath.Match(pattern, rel); m {
				return nil
			}
		}
	}

	line, err := w.cfg.mode.format(w.root, fullpath, dirent)
	if err != nil {
		return fmt.Errorf("can't format %s: %s", fullpath, err)
	}
	if _, err = w.bufw.WriteString(line); err != nil {
		return err
	}

	// Write path
	if _, err = w.bufw.WriteString(rel); err != nil {
		return err
	}
	return w.bufw.WriteByte('\n')
}

// Write prints the directory tree rooted at root in w.
func Write(w io.Writer, root string, opts ...Option) error {
	dtw, err := newWriter(w, root, opts...)
	if err != nil {
		return err
	}

	if err := filepath.WalkDir(root, dtw.walk); err != nil {
		return fmt.Errorf("dirtree: error walking directory: %v", err)
	}

	if err := dtw.bufw.Flush(); err != nil {
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
