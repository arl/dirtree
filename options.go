package dirtree

import (
	"fmt"
	"path/filepath"
)

type config struct {
	mode     PrintMode
	showRoot bool
	globs    []pattern
	depth    int
	types    filetype
}

var defaultCfg = config{
	mode:     ModeDefault,
	showRoot: true,
	globs:    nil,
	depth:    int(infiniteDepth),
	types:    typeFile | typeDir | typeOther,
}

// Option is the interface implemented by dirtree types used to control what to
// list and how to list it.
type Option interface {
	apply(*config) error
}

// A PrintMode represents the amount of information to print about a file, next
// to its filename. PrintMode is a bit set.
// Somewhat related to os.FileMode and fs.FileMode but much less detailed.
type PrintMode uint32

// implements the Option interface.
func (m PrintMode) apply(cfg *config) error {
	cfg.mode = m
	return nil
}

// The Type option limits the files to list based their type.
// Type can be formed of one or more of:
//  'f' for regular files
//  'd' for directories
//  '?' for anything else (symlink, etc.)
type Type string

func (t Type) apply(cfg *config) error {
	if t == "" {
		return fmt.Errorf("invalid Type: at least one type must be listed")
	}

	var types filetype
	for _, r := range string(t) {
		switch r {
		case rune(typeFile.char()):
			types |= typeFile
		case rune(typeDir.char()):
			types |= typeDir
		case rune(typeOther.char()):
			types |= typeOther
		default:
			return fmt.Errorf("invalid Type char %c, must be %c, %c or %c", r, typeFile.char(), typeDir.char(), typeOther.char())
		}
	}
	cfg.types = types
	return nil
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

type pattern struct {
	pat     string // pattern matched against
	discard bool   // indicates whether the file is discarded if it matches the pattern
}

func shouldKeepPath(path string, ps []pattern) bool {
	if ps == nil {
		return true
	}

	// Ignore patterns
	for _, p := range ps {
		match, _ := filepath.Match(p.pat, path)
		if match && p.discard {
			return false
		}
	}
	return true
}

// The Ignore option allows to ignore files matching a pattern. The path
// relative to the chosen root is matched against the pattern. Ignore follows
// the syntax used and described with the filepath.Match function. Before
// checking if it matches a pattern, a path is first converted to its slash
// ('/') based version, to ensure cross-platform consistency of the dirtree
// package.
//
// Ignore can be provided multiple times to ignore multiple patterns. A file is
// then ignored as soon as its path matches one pattern.
type Ignore string

func (i Ignore) apply(cfg *config) error {
	if _, err := filepath.Match(string(i), "/"); err != nil {
		return fmt.Errorf("invalid ignore pattern %v: %v", i, err)
	}
	cfg.globs = append(cfg.globs, pattern{pat: string(i), discard: true})
	return nil
}

// The Depth option indicates how many levels of directories below root should
// we recurse into. 0, the default, means there's no limit.
type Depth int

func (d Depth) apply(cfg *config) error {
	if d < 0 {
		return fmt.Errorf("negative Depth is invalid")
	}
	cfg.depth = int(d)
	return nil
}

const infiniteDepth Depth = 0
