package dirtree

import (
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

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

	// ModeDefault is a mask showing file type and size.
	ModeDefault PrintMode = ModeType | ModeSize

	// ModeAll is a mask showing all information about a file.
	ModeAll PrintMode = ModeType | ModeSize | ModeCRC32
)

type filetype byte

const (
	typeFile  filetype = 1 << iota // a regular file
	typeDir                        // a directory
	typeOther                      // anything else (symlink, whatever, ...)
)

// byte returns the printable char corresponding to ft.
func (ft filetype) char() byte {
	switch ft {
	case typeDir:
		return 'd'
	case typeFile:
		return 'f'
	case typeOther:
		return '?'
	}
	panic(fmt.Sprintf("filetype.Char(): unexpected filetype value: %d", ft))
}

func ftype(dirent fs.DirEntry) filetype {
	typ := dirent.Type()
	if typ.IsRegular() {
		return typeFile
	}
	if typ.Type() == fs.ModeDir {
		return typeDir
	}
	return typeOther
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

func checksum(fsys fs.FS, path string) (chksum string) {
	defer func() {
		if e := recover(); e != nil || chksum == "" {
			chksum = checksumNA()
		}
	}()
	var (
		f   fs.File
		err error
	)
	if fsys != nil {
		f, err = fsys.Open(path)
	} else {
		f, err = os.Open(path)
	}
	if err != nil {
		panic(err)
	}

	h := crc32.NewIEEE()
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

// format returns the file at fullpath, following the current print mode.
func (mode PrintMode) format(fsys fs.FS, fullpath string, dirent fs.DirEntry) (format string, err error) {
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
		sb.WriteByte(ft.char())
	}

	if mode&ModeSize != 0 {
		sep()

		var fi fs.FileInfo
		if fsys == nil {
			fi, err = os.Stat(fullpath)
		} else {
			fi, err = fs.Stat(fsys, fullpath)
		}
		if err != nil {
			return "", fmt.Errorf("failed to get size of %v: %v", fullpath, err)
		}

		sb.WriteString(formatSize(ft, fi.Size()))
	}

	if mode&ModeCRC32 != 0 {
		sep()
		sb.WriteString("crc=")
		if ft != typeFile {
			sb.WriteString(checksumNA())
		} else {
			sb.WriteString(checksum(fsys, fullpath))
		}
	}

	// Add a separator (if necessary)
	sep()
	return sb.String(), nil
}
