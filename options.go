package dirtree

import (
	"fmt"
	"path/filepath"
)

type config struct {
	mode     PrintMode
	showRoot bool
	ignore   []string
	depth    int
}

var defaultCfg = config{
	mode:     ModeDefault,
	showRoot: true,
	ignore:   nil,
	depth:    int(infiniteDepth),
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
