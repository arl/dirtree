package dirtree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// A PrintMode represents the amount of information to print about a file, next
// to its filename. PrintMode is a bit set.
// Somewhat related to os.FileMode and fs.FileMode but much less detailed.
type PrintMode uint32

const (
	// ModeType indicates if file is a directory, a file or something else.
	// It prints "t=dir", "t=file" or "t=other".
	ModeType PrintMode = 1 << iota

	// ModeSize reports the length in bytes for regular files. For other types
	// it shows NA (not applicable) since the size would be system dependent.
	// It prints "s=1234" or "s=NA".
	ModeSize

	// ModeSymlink indicates if a file is a symlink.
	// It prints "sym=true" or "sym=false".
	ModeSymlink

	// ModePerm reports unix permission bits.
	// It prints "perm=o644" for example.
	ModePerm

	// ModeAll is a mask showing all information aout a file.
	ModeAll PrintMode = ModeType | ModeSize | ModeSymlink | ModePerm

	// ModeStd is a mask showing kind and size all standard information aout a file. Should be
	// enough in most cases.
	ModeStd PrintMode = ModeType | ModeSize
)

type ftype int

const (
	typeDir ftype = iota
	typeFile
	typeOther
)

func filetype(dirent fs.DirEntry) ftype {
	switch {
	case dirent.Type().IsDir():
		return typeDir
	case dirent.Type().IsRegular():
		return typeFile
	default:
		return typeOther
	}
}

func (t ftype) String() string {
	switch t {
	case typeDir:
		return "dir"
	case typeFile:
		return "file"
	case typeOther:
		return "other"
	}

	panic(fmt.Sprintf("invalid filetype ftype(%d)", t))
}

// format prints the name
func (mode PrintMode) format(root, fullpath string, dirent fs.DirEntry) (format string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	var sb strings.Builder
	rel, err := filepath.Rel(root, fullpath)
	if err != nil {
		return "", fmt.Errorf("can't find relative path: %s", err)
	}

	var fi fs.FileInfo // lazy
	stat := func() fs.FileInfo {
		if fi != nil {
			return fi
		}
		fi, err = os.Lstat(fullpath)
		if err != nil {
			panic(fmt.Errorf("can't get size: %v", err))
		}
		return fi
	}

	ft := filetype(dirent)

	sb.WriteString(rel)
	if mode&ModeType != 0 {
		sb.WriteByte(' ')
		sb.WriteString("t=")
		sb.WriteString(ft.String())
	}

	if mode&ModeSize != 0 {
		sb.WriteByte(' ')
		sb.WriteString("s=")
		switch ft {
		case typeFile:
			fmt.Fprintf(&sb, "%d", stat().Size())
		default:
			sb.WriteByte('0')
		}
	}

	return sb.String(), err
}
