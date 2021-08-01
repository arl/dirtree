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
	root := filepath.Join("testdata", "dir")
	dirA := filepath.Join(root, "A")
	file1 := filepath.Join(root, "A", "file1")
	symfile1 := filepath.Join(root, "A", "symfile1")
	symdirA := filepath.Join(root, "A", "B", "symdirA")

	tests := []struct {
		name     string
		mode     PrintMode
		root     string
		fullpath string
		dirent   fs.DirEntry
		want     string
		wantErr  bool
	}{
		{
			name: "mode=ModeType/file1",
			mode: ModeType,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: "f ",
		},
		{
			name: "mode=ModeSize/file1",
			mode: ModeSize,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: "13b        ",
		},
		{
			name: "mode=ModeStd/file1",
			mode: ModeStd,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: "f 13b        ",
		},
		{
			name: "mode=ModeAll/file1",
			mode: ModeAll,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: "f 13b        crc=0451ac5e ",
		},
		{
			name: "mode=ModeStd/dirA",
			mode: ModeStd,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			want: "d            ",
		},
		{
			name: "mode=ModeType/symfile1",
			mode: ModeStd,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			want: "?            ",
		},
		{
			name: "mode=ModeType/symdirA",
			mode: ModeStd,
			root: root, fullpath: symdirA, dirent: newDentry(symdirA),
			want: "?            ",
		},
		{
			name: "mode=ModeCRC32/file1",
			mode: ModeCRC32,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: "crc=0451ac5e ",
		},
		{
			name: "mode=ModeCRC32/dirA",
			mode: ModeCRC32,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			want: "crc=n/a      ",
		},
		{
			name: "mode=ModeCRC32/symfile1",
			mode: ModeCRC32,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			want: "crc=n/a      ",
		},

		// Error cases
		{
			name: "mode=ModeAll/do-not-exist",
			mode: ModeAll,
			root: root, fullpath: "do-not-exist", dirent: newDentry(symdirA),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mode.format(tt.root, tt.fullpath, tt.dirent)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrintMode.format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("format error\ngot :%q\nwant:%q", got, tt.want)
			}
		})
	}
}

func Test_checksumNA(t *testing.T) {
	// Verify that checksum does not fail on error and that instead, it returns
	// the string returned by checksumNA. Errors are caught before.
	got := checksum(typeFile, "do-not-exist")
	if got != checksumNA() {
		t.Errorf("checksum() = %v, want %v", got, checksumNA())
	}
}
