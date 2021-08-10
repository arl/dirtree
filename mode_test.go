package dirtree

import (
	"path/filepath"
	"testing"
	"testing/fstest"
)

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
		ft       filetype
		want     string
		wantErr  bool
	}{
		{
			name: "mode=ModeType/file1",
			mode: ModeType,
			root: root, fullpath: file1, ft: typeFile,
			want: "f ",
		},
		{
			name: "mode=ModeSize/file1",
			mode: ModeSize,
			root: root, fullpath: file1, ft: typeFile,
			want: "13b        ",
		},
		{
			name: "mode=ModeStd/file1",
			mode: ModeDefault,
			root: root, fullpath: file1, ft: typeFile,
			want: "f 13b        ",
		},
		{
			name: "mode=ModeAll/file1",
			mode: ModeAll,
			root: root, fullpath: file1, ft: typeFile,
			want: "f 13b        crc=0451ac5e ",
		},
		{
			name: "mode=ModeStd/dirA",
			mode: ModeDefault,
			root: root, fullpath: dirA, ft: typeDir,
			want: "d            ",
		},
		{
			name: "mode=ModeType/symfile1",
			mode: ModeDefault,
			root: root, fullpath: symfile1, ft: typeOther,
			want: "?            ",
		},
		{
			name: "mode=ModeType/symdirA",
			mode: ModeDefault,
			root: root, fullpath: symdirA, ft: typeOther,
			want: "?            ",
		},
		{
			name: "mode=ModeCRC32/file1",
			mode: ModeCRC32,
			root: root, fullpath: file1, ft: typeFile,
			want: "crc=0451ac5e ",
		},
		{
			name: "mode=ModeCRC32/dirA",
			mode: ModeCRC32,
			root: root, fullpath: dirA, ft: typeDir,
			want: "crc=n/a      ",
		},
		{
			name: "mode=ModeCRC32/symfile1",
			mode: ModeCRC32,
			root: root, fullpath: symfile1, ft: typeOther,
			want: "crc=n/a      ",
		},

		// Error cases
		{
			name: "mode=ModeAll/do-not-exist",
			mode: ModeAll,
			root: root, fullpath: "do-not-exist", ft: typeOther,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mode.format(nil, tt.fullpath, tt.ft)
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
	t.Run("fsys=nil", func(t *testing.T) {
		if got := checksum(nil, "do-not-exist"); got != checksumNA() {
			t.Errorf("checksum() = %v, want %v", got, checksumNA())
		}
	})
	t.Run("fsys=MapFS", func(t *testing.T) {
		if got := checksum(fstest.MapFS{}, "do-not-exist"); got != checksumNA() {
			t.Errorf("checksum() = %v, want %v", got, checksumNA())
		}
	})
}
