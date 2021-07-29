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
	ignore   []string
	depth    int
}

var defaultCfg = config{
	mode:     ModeAll,
	showRoot: true,
	ignore:   nil,
	depth:    int(infiniteDepth),
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

// The Ignore option defines a pattern allowing to ignore certain files to be
// printed, depending on their relative path, with respect to the chosen root.
// Ignore follows the syntax used and described with the filepath.Match
// function. Before checking if it matches a pattern, a path is first converted
// to its slash ('/') based version, to ensure cross-platform consistency of the
// dirtree package.
// Ignore can be provided multiple times to ignore multiple patterns.
type Ignore string

func (i Ignore) apply(cfg *config) error {
	if _, err := filepath.Match(string(i), "/"); err != nil {
		return fmt.Errorf("invalid ignore pattern %v: %v", i, err)
	}
	cfg.ignore = append(cfg.ignore, string(i))
	return nil
}

type Depth int

func (d Depth) apply(cfg *config) error {
	if d < 0 {
		return fmt.Errorf("negative Depth is invalid")
	}
	cfg.depth = int(d)
	return nil
}

const infiniteDepth Depth = 0

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
