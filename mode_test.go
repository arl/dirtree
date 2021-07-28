package dirtree

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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

// Create the following directory structure rooted at 'root':
//	.
//	└── A
//		├── B
//		│   └── symdirA -> symlink to A
//		├── file1
//		└── symfile1 -> symlink to A/file1
func createDirStructure(tb testing.TB, root string) (dirA, file1, symfile1, symdirA string) {
	if err := os.Chmod(root, 0o744); err != nil {
		tb.Fatal(err)
	}

	dirA = filepath.Join(root, "A")
	file1 = filepath.Join(root, "A", "file1")
	symfile1 = filepath.Join(root, "A", "symfile1")
	symdirA = filepath.Join(root, "A", "B", "symdirA")

	if err := os.MkdirAll(filepath.Join(dirA, "B"), 0o744); err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(file1, []byte("dummy content"), 0o744); err != nil {
		tb.Fatal(err)
	}
	if err := os.Symlink(file1, symfile1); err != nil {
		tb.Fatal(err)
	}
	if err := os.Symlink(dirA, symdirA); err != nil {
		tb.Fatal(err)
	}

	return
}

func TestPrintMode_format(t *testing.T) {
	root := t.TempDir()
	dirA, file1, symfile1, symdirA := createDirStructure(t, root)

	tests := []struct {
		name     string
		mode     PrintMode
		root     string
		fullpath string
		dirent   fs.DirEntry
		want     map[string]string
		wantErr  bool
	}{
		{
			name: "mode=ModeType/file1",
			mode: ModeType,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: map[string]string{"all": "f A/file1"},
		},
		{
			name: "mode=ModeSize/file1",
			mode: ModeSize,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: map[string]string{"all": "13b        A/file1"},
		},
		{
			name: "mode=ModeStd/file1",
			mode: ModeStd,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: map[string]string{"all": "f 13b        A/file1"},
		},
		{
			name: "mode=ModeAll/file1",
			mode: ModeAll,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: map[string]string{"all": "f 13b        crc=0451ac5e A/file1"},
		},
		{
			name: "mode=ModeStd/dirA",
			mode: ModeStd,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			want: map[string]string{"all": "d            A"},
		},
		{
			name: "mode=ModeType/symfile1",
			mode: ModeStd,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			want: map[string]string{"all": "?            A/symfile1"},
		},
		{
			name: "mode=ModeType/symdirA",
			mode: ModeStd,
			root: root, fullpath: symdirA, dirent: newDentry(symdirA),
			want: map[string]string{"all": "?            A/B/symdirA"},
		},
		{
			name: "mode=ModeCRC32/file1",
			mode: ModeCRC32,
			root: root, fullpath: file1, dirent: newDentry(file1),
			want: map[string]string{"all": "crc=0451ac5e A/file1"},
		},
		{
			name: "mode=ModeCRC32/dirA",
			mode: ModeCRC32,
			root: root, fullpath: dirA, dirent: newDentry(dirA),
			want: map[string]string{"all": "crc=n/a      A"},
		},
		{
			name: "mode=ModeCRC32/symfile1",
			mode: ModeCRC32,
			root: root, fullpath: symfile1, dirent: newDentry(symfile1),
			want: map[string]string{"all": "crc=n/a      A/symfile1"},
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
			want, ok := tt.want["all"]
			if !ok {
				// Platform dependent test case.
				want, ok = tt.want[runtime.GOOS]
				if !ok {
					t.Skipf("Case not tested yet on GOOS=%v, please add format an open a pull-request!", runtime.GOOS)
				}
			}
			if got != want {
				t.Errorf("format error\ngot :%q\nwant:%q", got, want)
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
