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
	types    FileType
}

var defaultCfg = config{
	mode:     ModeDefault,
	showRoot: true,
	globs:    nil,
	depth:    int(infiniteDepth),
	types:    File | Dir | Other,
}

// Option is the interface implemented by dirtree types used to control what to
// list and how to list it.
type Option interface {
	apply(*config) error
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

	var types FileType
	for _, r := range string(t) {
		switch r {
		case rune(File.char()):
			types |= File
		case rune(Dir.char()):
			types |= Dir
		case rune(Other.char()):
			types |= Other
		default:
			return fmt.Errorf("invalid Type char %c, must be %c, %c or %c", r, File.char(), Dir.char(), Other.char())
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
	pat string        // pattern matched against
	moi matchOrIgnore // is this a match or an ignore pattern
}

func shouldKeepPath(path string, ps []pattern) bool {
	if ps == nil {
		return true
	}

	keep := false
	hasMatch := false
	for _, p := range ps {
		m, _ := filepath.Match(p.pat, path)
		if m && p.moi == ignore {
			return false
		}
		if p.moi == match {
			hasMatch = true
			keep = keep || m
		}
	}
	return !hasMatch || keep
}

// The Ignore option allows to ignore files matching a pattern. The path
// relative to the chosen root is matched against the pattern. Ignore follows
// the syntax used and described with the filepath.Match function. Before
// checking if it matches a pattern, a path is first converted to its slash
// ('/') based version, to ensure cross-platform consistency of the dirtree
// package.
//
// Ignore can be provided multiple times to ignore multiple patterns. A file is
// ignored from the listing as long as at it matches at least one Ignore
// pattern. Also, Ignore has precedence over Match.
type Ignore string

func (i Ignore) apply(cfg *config) error {
	if _, err := filepath.Match(string(i), "/"); err != nil {
		return fmt.Errorf("invalid ignore pattern %v: %v", i, err)
	}
	cfg.globs = append(cfg.globs, pattern{pat: string(i), moi: ignore})
	return nil
}

// The Match option limits the listing to files that match a pattern. The path
// relative to the chosen root is matched against the pattern. Match follows the
// syntax used and described with the filepath.Match function. Before checking
// if it matches a pattern, a path is first converted to its slash ('/') based
// version, to ensure cross-platform consistency of the dirtree package.
//
// Match can be provided multiple times to match multiple patterns. A file is
// included in the listing as long as at it matches at least one Match pattern,
// unless it matches an Ignore pattern (since Ignore has precedence over Match).
type Match string

func (m Match) apply(cfg *config) error {
	if _, err := filepath.Match(string(m), "/"); err != nil {
		return fmt.Errorf("invalid match pattern %v: %v", m, err)
	}
	cfg.globs = append(cfg.globs, pattern{pat: string(m), moi: match})
	return nil
}

type matchOrIgnore bool

const (
	match  matchOrIgnore = true
	ignore matchOrIgnore = false
)

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
