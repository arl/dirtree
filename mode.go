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

// A PrintMode represents the amount of information to print about a file, next
// to its filename. PrintMode is a bit set.
// Somewhat related to os.FileMode and fs.FileMode but much less detailed.
type PrintMode uint32

// implements the Option interface.
func (m PrintMode) apply(cfg *config) error {
	cfg.mode = m
	return nil
}

type FileType byte

const (
	File  FileType = 1 << iota // File is for regular files
	Dir                        // Dir is for directories
	Other                      // Other is for anything else (symlink, whatever, ...)
)

// byte returns the printable char corresponding to ft.
func (ft FileType) char() byte {
	switch ft {
	case Dir:
		return 'd'
	case File:
		return 'f'
	case Other:
		return '?'
	}
	panic(fmt.Sprintf("FileType.Char(): unexpected FileType value: %d", ft))
}

func filetypeFromDirEntry(dirent fs.DirEntry) FileType {
	typ := dirent.Type()
	if typ.IsRegular() {
		return File
	}
	if typ.Type() == fs.ModeDir {
		return Dir
	}
	return Other
}

// we pad the size to sizeDigits, with spaces, so that for most filenames all
// the fields are aligned. We're hoewever not going to truncate the size of bigger files
// just to respect that rule, we're making an exception in those cases.
const sizeDigits = 9

func formatSize(ft FileType, size int64) string {
	if ft != File {
		return fmt.Sprintf("%-*s", sizeDigits+1, "")
	}
	str := strconv.FormatInt(size, 10) + "b"
	if len(str) > sizeDigits {
		return str
	}

	return fmt.Sprintf("%-*s", sizeDigits+1, str)
}

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
	if _, err := io.Copy(h, f); err != nil {
		panic(err)
	}

	chksum = fmt.Sprintf("%0*x", crcChars, h.Sum32())
	return
}

const na = "n/a"

func checksumNA() string {
	return fmt.Sprintf("%-*s", crcChars, na)
}

// An Entry holds gathered information about a particular file.
type Entry struct {
	Path     string
	Type     FileType
	Size     int64
	Checksum string

	mode PrintMode
}

func newEntry(mode PrintMode, fsys fs.FS, fullpath string, ft FileType) (*Entry, error) {
	ent := &Entry{
		mode: mode,
		Type: ft,
	}

	if mode&ModeSize != 0 {
		var (
			fi  fs.FileInfo
			err error
		)
		if fsys == nil {
			fi, err = os.Stat(fullpath)
		} else {
			fi, err = fs.Stat(fsys, fullpath)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get size of %v: %v", fullpath, err)
		}
		ent.Size = fi.Size()
	}

	if mode&ModeCRC32 != 0 {
		if ft != File {
			ent.Checksum = na
		} else {
			ent.Checksum = checksum(fsys, fullpath)
		}
	}

	return ent, nil
}

// Format returns a summary string of e. Some information might be missing,
// depending on the PrintMode used to create the Entry.
func (e *Entry) Format() string {
	var sb strings.Builder

	// Separate successive mode expressions
	sep := func() {
		if sb.Len() != 0 {
			sb.WriteByte(' ')
		}
	}

	if e.mode&ModeType != 0 {
		sep()
		sb.WriteByte(e.Type.char())
	}

	if e.mode&ModeSize != 0 {
		sep()
		sb.WriteString(formatSize(e.Type, e.Size))
	}

	if e.mode&ModeCRC32 != 0 {
		sep()
		sb.WriteString("crc=")
		if e.Type != File {
			sb.WriteString(checksumNA())
		} else {
			sb.WriteString(e.Checksum)
		}
	}

	// Add a separator (if necessary)
	sep()
	return sb.String()
}
