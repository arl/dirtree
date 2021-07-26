package dirtree

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

type dentry struct {
	info fs.FileInfo
}

func newDentry(path string) *dentry {
	info, err := os.Lstat(path)
	if err != nil {
		panic(err)
	}
	return &dentry{info}
}

func (d *dentry) Name() string               { return d.info.Name() }
func (d *dentry) IsDir() bool                { return d.info.IsDir() }
func (d *dentry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d *dentry) Info() (fs.FileInfo, error) { return d.info, nil }

func TestPrintMode_format(t *testing.T) {
	root := t.TempDir()

	// Create a directory structure rooted at 'root'.
	// TODO(arl) add structure with 'tree' output

	var (
		dirA     = filepath.Join(root, "A")
		file1    = filepath.Join(root, "A", "file1")
		symfile1 = filepath.Join(root, "A", "symfile1")
		symdirA  = filepath.Join(root, "A", "B", "symdirA")
	)
	if err := os.MkdirAll(filepath.Join(dirA, "B"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file1, []byte("dummy content"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(file1, symfile1); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(dirA, symdirA); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		mode       PrintMode
		root       string
		fullpath   string
		dirent     fs.DirEntry
		wantFormat string
		wantErr    bool
	}{
		{
			name: "mode=ModeType/file1",
			mode: ModeType,
			root: root, fullpath: file1, dirent: newDentry(file1),
			wantFormat: "f A/file1",
		},
		{
			name: "mode=ModeSize/file1",
			mode: ModeSize,
			root: root, fullpath: file1, dirent: newDentry(file1),
			wantFormat: "13b        A/file1",
		},
		{
			name: "mode=ModeStd/file1",
			mode: ModeStd,
			root: root, fullpath: file1, dirent: newDentry(file1),
			wantFormat: "f 13b        A/file1",
		},
		{
			name: "mode=ModeAll/file1",
			mode: ModeAll,
			root: root, fullpath: file1, dirent: newDentry(file1),
			wantFormat: "f 775 sym=0 13b        crc=0451ac5e A/file1",
		},
		{
			name: "mode=ModeStd+ModeSymlink/dirA",
			mode: ModeStd | ModeSymlink,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			wantFormat: "d sym=0            A",
		},
		{
			name: "mode=ModeType+ModeSymlink/symfile1",
			mode: ModeStd | ModeSymlink,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			wantFormat: "? sym=1            A/symfile1",
		},
		{
			name: "mode=ModeType+ModeSymlink/symdirA",
			mode: ModeStd | ModeSymlink,
			root: root, fullpath: symdirA, dirent: newDentry(symdirA),
			wantFormat: "? sym=1            A/B/symdirA",
		},
		{
			name: "mode=ModeCRC32/file1",
			mode: ModeCRC32,
			root: root, fullpath: file1, dirent: newDentry(file1),
			wantFormat: "crc=0451ac5e A/file1",
		},
		{
			name: "mode=ModeCRC32/dirA",
			mode: ModeCRC32,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			wantFormat: "crc=n/a      A",
		},
		{
			name: "mode=ModeCRC32/symfile1",
			mode: ModeCRC32,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			wantFormat: "crc=n/a      A/symfile1",
		},
		{
			name: "mode=ModePerm/symdirA",
			mode: ModePerm,
			root: root, fullpath: symdirA, dirent: newDentry(symdirA),
			wantFormat: "777 A/B/symdirA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mode.format(tt.root, tt.fullpath, tt.dirent)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintMode.format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantFormat {
				t.Errorf("format error\ngot :%q\nwant:%q", got, tt.wantFormat)
			}
		})
	}
}

func Test_checksumENOENT(t *testing.T) {
	notExist := filepath.Join(t.TempDir(), "do-not-exist")
	got := checksum(typeFile, notExist)
	if got != checksumNA() {
		t.Errorf("checksum() = %v, want %v", got, checksumNA())
	}
}
