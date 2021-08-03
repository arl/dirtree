package dirtree

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

var tests = []struct {
	name    string
	opts    []Option
	want    []string
	wantErr bool
}{
	{
		name: "default",
		opts: nil,
		want: []string{
			"d            .",
			"d            A",
			"d            A/B",
			"?            A/B/symdirA",
			"f 13b        A/file1",
			"?            A/symfile1",
		},
	},
	{
		name: "all details",
		opts: []Option{ModeAll},
		want: []string{
			"d            crc=n/a      .",
			"d            crc=n/a      A",
			"d            crc=n/a      A/B",
			"?            crc=n/a      A/B/symdirA",
			"f 13b        crc=0451ac5e A/file1",
			"?            crc=n/a      A/symfile1",
		},
	},
	{
		name: "exclude root",
		opts: []Option{ExcludeRoot},
		want: []string{
			"d            A",
			"d            A/B",
			"?            A/B/symdirA",
			"f 13b        A/file1",
			"?            A/symfile1",
		},
	},
	{
		name: "include root and crc32",
		opts: []Option{IncludeRoot(true), ModeCRC32},
		want: []string{
			"crc=n/a      .",
			"crc=n/a      A",
			"crc=n/a      A/B",
			"crc=n/a      A/B/symdirA",
			"crc=0451ac5e A/file1",
			"crc=n/a      A/symfile1",
		},
	},
	{
		name: `ignore single`,
		opts: []Option{Ignore("*/file1")},
		want: []string{
			"d            .",
			"d            A",
			"d            A/B",
			"?            A/B/symdirA",
			"?            A/symfile1",
		},
	},
	{
		name: `ignore multiple`,
		opts: []Option{Ignore("*/file1"), Ignore("A")},
		want: []string{
			"d            .",
			"d            A/B",
			"?            A/B/symdirA",
			"?            A/symfile1",
		},
	},
	{
		name: `depth 1`,
		opts: []Option{ModeType, Depth(1)},
		want: []string{
			"d .",
			"d A",
		},
	},
	{
		name: `depth 2 and no root`,
		opts: []Option{ModeType, Depth(2), ExcludeRoot},
		want: []string{
			"d A",
			"d A/B",
			"f A/file1",
			"? A/symfile1",
		},
	},

	// Error cases
	{
		name:    "invalid ignore pattern",
		opts:    []Option{Ignore("a/b[")},
		wantErr: true,
	},
	{
		name:    "negative depth",
		opts:    []Option{Depth(-1)},
		wantErr: true,
	},
}

func TestSprint(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sprint(filepath.Join("testdata", "dir"), tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sprint() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			want := strings.Join(tt.want, "\n")
			if got = strings.TrimSpace(got); got != want {
				t.Errorf("Sprint, invalid output:\ngot:\n%v\n\nwant:\n%s", got, want)
			}
		})
	}
}

func TestSprintFS(t *testing.T) {
	fsys := fstest.MapFS{
		"A/file1":     &fstest.MapFile{Data: []byte("dummy content")},
		"A/symfile1":  &fstest.MapFile{Mode: fs.ModeSymlink},
		"A/B/symdirA": &fstest.MapFile{Mode: fs.ModeSymlink | fs.ModeDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SprintFS(fsys, ".", tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SprintFS() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			want := strings.Join(tt.want, "\n")
			if got = strings.TrimSpace(got); got != want {
				t.Errorf("SprintFS, invalid output:\ngot:\n%v\n\nwant:\n%s", got, want)
			}
		})
	}
}

func BenchmarkWrite(b *testing.B) {
	/*
		This benchmarks runs on a directory structure of 11110 directories and
		11110 files, filled with 1024byte of random data, created with:

		ulimit -S -n 20000
		cd $(mktemp -d)
		mkdir -p A{0,1,2,3,4,5,6,7,8,9}/B{0,1,2,3,4,5,6,7,8,9}/C{0,1,2,3,4,5,6,7,8,9}/D{0,1,2,3,4,5,6,7,8,9}
		head -c 1024 /dev/urandom | tee A{0,1,2,3,4,5,6,7,8,9}/B{0,1,2,3,4,5,6,7,8,9}/C{0,1,2,3,4,5,6,7,8,9}/D{0,1,2,3,4,5,6,7,8,9}/file > /dev/null
		head -c 1024 /dev/urandom | tee A{0,1,2,3,4,5,6,7,8,9}/B{0,1,2,3,4,5,6,7,8,9}/C{0,1,2,3,4,5,6,7,8,9}/file > /dev/null
		head -c 1024 /dev/urandom | tee A{0,1,2,3,4,5,6,7,8,9}/B{0,1,2,3,4,5,6,7,8,9}/file > /dev/null
		head -c 1024 /dev/urandom | tee A{0,1,2,3,4,5,6,7,8,9}/file > /dev/null
	*/

	const dir = "/tmp/tmp.YGmxHbsmj1"
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		Write(io.Discard, dir, Depth(2))
	}
}
