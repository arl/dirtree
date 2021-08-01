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

// WriteFS walks the directory rooted at root in the given filesystem and prints
// one file per line into w.
//
// A variable number of options can be provided to control the limit the files
// printed and/or the amount of information printed for each of them.
func WriteFS(w io.Writer, fsys fs.FS, root string, opts ...Option) error {
	return write(w, root, fsys, opts...)
}

// Write walks the directory rooted at root and prints one file per line into w.
//
// A variable number of options can be provided to control the limit the files
// printed and/or the amount of information printed for each of them.
func Write(w io.Writer, root string, opts ...Option) error {
	return WriteFS(w, nil, root, opts...)
}

// SprintFS walks the directory rooted at root in the given filesystem and
// returns the list of files.
//
// It's a wrapper around WriteFS(...) provided for convenience.
func SprintFS(fsys fs.FS, root string, opts ...Option) (string, error) {
	var sb strings.Builder
	if err := WriteFS(&sb, fsys, root, opts...); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Sprint walks the directory rooted at root and returns a string containing the
// list of files.
//
// It's a wrapper around Write(...) provided for convenience.
func Sprint(root string, opts ...Option) (string, error) {
	return SprintFS(nil, root, opts...)
}

// write walks through all files of fsys, starting at root. Use actual
// filesystem if fsys is nil.
func write(w io.Writer, root string, fsys fs.FS, opts ...Option) error {
	// Configure the walk
	cfg := defaultCfg
	for _, o := range opts {
		if err := o.apply(&cfg); err != nil {
			return fmt.Errorf("dirtree: configuration error: %v", err)
		}
	}

	walkdir := fs.WalkDir
	seenRoot := false
	bufw := bufio.NewWriter(w)

	if fsys == nil {
		walkdir = func(_ fs.FS, root string, fn fs.WalkDirFunc) error {
			return filepath.WalkDir(root, fn)
		}
	}

	// Do walk
	walk := func(fullpath string, dirent fs.DirEntry, err error) error {
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

		// Path conversion: relative to root and slash based
		rel, err := filepath.Rel(root, fullpath)
		if err != nil {
			return err
		}

		// Depth check
		if cfg.depth != 0 {
			if len(strings.Split(rel, string(os.PathSeparator))) > cfg.depth {
				if dirent.IsDir() {
					err = fs.SkipDir
				}
				return err
			}
		}

		rel = filepath.ToSlash(rel)

		// Ignore patterns
		if cfg.ignore != nil {
			for _, pattern := range cfg.ignore {
				if m, _ := filepath.Match(pattern, rel); m {
					return nil
				}
			}
		}

		line, err := cfg.mode.format(fsys, fullpath, dirent)
		if err != nil {
			return fmt.Errorf("can't format %s: %s", fullpath, err)
		}
		if _, err = bufw.WriteString(line); err != nil {
			return err
		}

		// Write path
		if _, err = bufw.WriteString(rel); err != nil {
			return err
		}
		return bufw.WriteByte('\n')
	}

	if err := walkdir(fsys, root, walk); err != nil {
		return fmt.Errorf("dirtree: error walking directory: %v", err)
	}

	if err := bufw.Flush(); err != nil {
		return fmt.Errorf("dirtree: can't write output: %s", err)
	}
	return nil

}
