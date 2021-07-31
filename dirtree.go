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

// Write prints the directory tree rooted at root in w.
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

		line, err := cfg.mode.format(root, fullpath, dirent)
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
