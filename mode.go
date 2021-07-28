package dirtree

import (
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// A PrintMode represents the amount of information to print about a file, next
// to its filename. PrintMode is a bit set.
// Somewhat related to os.FileMode and fs.FileMode but much less detailed.
type PrintMode uint32

const (
	// ModeType indicates if file is a directory, a regular file or something
	// else. It prints 'd', 'f' or '?' respectively.
	ModeType PrintMode = 1 << iota

	// ModeSize reports the length in bytes for regular files, "1234b" for
	// example, or nothing for other types where size is not applicable (it
	// would be OS-dependent).
	ModeSize

	// ModeCRC32 computes and reports the CRC-32 checksum for regular files. For
	// other file types, or for files which permissions prevent reading, it
	// shows n/a (i.e. not applicable). Example "crc=294a245b" or "crc=n/a"
	ModeCRC32

	// ModeStd is a mask showing file type and size.
	ModeStd PrintMode = ModeType | ModeSize

	// ModeAll is a mask showing all information about a file.
	ModeAll PrintMode = ModeType | ModeSize | ModeCRC32
)

type filetype byte

const (
	typeDir   filetype = 'd'
	typeFile  filetype = 'f'
	typeOther filetype = '?'
)

func ftype(dirent fs.DirEntry) filetype {
	switch {
	case dirent.Type().IsDir():
		return typeDir
	case dirent.Type().IsRegular():
		return typeFile
	default:
		return typeOther
	}
}

// we pad the size to sizeDigits, with spaces, so that for most filenames all
// the fields are aligned. We're hoewever not going to truncate the size of bigger files
// just to respect that rule, we're making an exception in those cases.
const sizeDigits = 9

func formatSize(ft filetype, size int64) string {
	if ft != typeFile {
		return fmt.Sprintf("%-*s", sizeDigits+1, "")
	}
	str := strconv.FormatInt(size, 10) + "b"
	if len(str) > sizeDigits {
		return str
	}

	return fmt.Sprintf("%-*s", sizeDigits+1, str)
}

// buffer used in io.CopyBUffer to reduce allocations
// while calculating the file checksum.
var iobuf [32 * 1024]byte

// number of chars in hexadecimal representation of a CRC-32.
const crcChars = crc32.Size * 2 // 2 since 2 chars per raw byte

func checksum(ft filetype, path string) (chksum string) {
	defer func() {
		if e := recover(); e != nil || chksum == "" {
			chksum = checksumNA()
		}
	}()

	if ft != typeFile {
		return
	}

	h := crc32.NewIEEE()
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.CopyBuffer(h, f, iobuf[:]); err != nil {
		panic(err)
	}

	chksum = fmt.Sprintf("%0*x", crcChars, h.Sum32())
	return
}

func checksumNA() string {
	const na = "n/a"
	return fmt.Sprintf("%-*s", crcChars, na)
}

// format returns the file at fullpath, roots it at root, following the current print mode.
func (mode PrintMode) format(root, fullpath string, dirent fs.DirEntry) (format string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	var sb strings.Builder
	ft := ftype(dirent)

	// Separate successive mode expressions
	sep := func() {
		if sb.Len() != 0 {
			sb.WriteByte(' ')
		}
	}

	if mode&ModeType != 0 {
		sep()
		sb.WriteByte(byte(ft))
	}

	if mode&ModeSize != 0 {
		sep()

		fi, err := os.Lstat(fullpath)
		if err != nil {
			return "", fmt.Errorf("failed to get size of %v: %v", fullpath, err)
		}

		sb.WriteString(formatSize(ft, fi.Size()))
	}

	if mode&ModeCRC32 != 0 {
		sep()
		sb.WriteString("crc=")
		sb.WriteString(checksum(ft, fullpath))
	}

	// Write file name
	sep()
	rel, err := filepath.Rel(root, fullpath)
	if err != nil {
		return "", fmt.Errorf("can't find relative path: %s", err)
	}

	// Convert to slash-based path.
	sb.WriteString(filepath.ToSlash(rel))
	return sb.String(), err
}
