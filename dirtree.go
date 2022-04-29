// Package dirtree implements utility routines to list files in a deterministic
// and cross-patform manner.
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

// List walks the directory rooted at root in the given filesystem and returns
// entries.
//
// A variable number of options can be provided to control the limit the files
// printed and/or the amount of information gathered for each of them.
func List(fsys fs.FS, root string, opts ...Option) ([]*Entry, error) {
	entries, err := walkTree(root, fsys, opts...)
	if err != nil {
		return nil, fmt.Errorf("dirtree: %v", err)
	}
	return entries, nil
}

// WriteFS walks the directory rooted at root in the given filesystem and prints
// one file per line into w.
//
// A variable number of options can be provided to control the limit the files
// printed and/or the amount of information printed for each of them.
func WriteFS(w io.Writer, fsys fs.FS, root string, opts ...Option) error {
	entries, err := walkTree(root, fsys, opts...)
	if err != nil {
		return fmt.Errorf("dirtree: %v", err)
	}
	if err := writeEntries(w, entries); err != nil {
		return fmt.Errorf("dirtree: %v", err)
	}
	return nil
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

func writeEntries(w io.Writer, entries []*Entry) error {
	bufw := bufio.NewWriter(w)

	for _, ent := range entries {
		if _, err := bufw.WriteString(ent.Format()); err != nil {
			return err
		}

		// Write path
		if _, err := bufw.WriteString(ent.RelPath); err != nil {
			return err
		}
		bufw.WriteByte('\n')
	}

	if err := bufw.Flush(); err != nil {
		return fmt.Errorf("can't write output: %s", err)
	}
	return nil
}

// walkTree walks through all files of fsys, starting at root, and returns the
// files, in the order they're met, as entries. Use actual filesystem if fsys is
// nil.
func walkTree(root string, fsys fs.FS, opts ...Option) ([]*Entry, error) {
	// Configure the walk
	cfg := defaultCfg
	for _, o := range opts {
		if err := o.apply(&cfg); err != nil {
			return nil, fmt.Errorf("configuration error: %v", err)
		}
	}

	walkdir := fs.WalkDir
	seenRoot := false

	if fsys == nil {
		walkdir = func(_ fs.FS, root string, fn fs.WalkDirFunc) error {
			return filepath.WalkDir(root, fn)
		}
	}

	entries := make([]*Entry, 0, 128)
	// Do walk
	walk := func(fullpath string, dirent fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip based on type
		ft := filetypeFromDirEntry(dirent)
		if cfg.types&ft == 0 {
			return nil
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
		if !shouldKeepPath(rel, cfg.globs) {
			return nil
		}

		ent, err := newEntry(cfg.mode, fsys, fullpath, ft)
		if err != nil {
			return fmt.Errorf("can't create Entry for %s: %s", fullpath, err)
		}
		ent.RelPath = rel

		entries = append(entries, ent)
		return nil
	}

	if err := walkdir(fsys, root, walk); err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}
	return entries, nil
}
