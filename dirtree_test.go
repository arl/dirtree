package dirtree

import (
	"strings"
	"testing"
)

func TestSprint(t *testing.T) {
	root := t.TempDir()
	createDirStructure(t, root)

	tests := []struct {
		name    string
		opts    []Option
		want    []string
		wantErr bool
	}{
		{
			name: "default",
			opts: nil,
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
				"d            crc=n/a      A",
				"d            crc=n/a      A/B",
				"?            crc=n/a      A/B/symdirA",
				"f 13b        crc=0451ac5e A/file1",
				"?            crc=n/a      A/symfile1",
			},
		},
		{
			name: "include root",
			opts: []Option{IncludeRoot(true)},
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
			name: `ignore single`,
			opts: []Option{Ignore("*/file1")},
			want: []string{
				"d            crc=n/a      .",
				"d            crc=n/a      A",
				"d            crc=n/a      A/B",
				"?            crc=n/a      A/B/symdirA",
				"?            crc=n/a      A/symfile1",
			},
		},
		{
			name: `ignore multiple`,
			opts: []Option{Ignore("*/file1"), Ignore("A")},
			want: []string{
				"d            crc=n/a      .",
				"d            crc=n/a      A/B",
				"?            crc=n/a      A/B/symdirA",
				"?            crc=n/a      A/symfile1",
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sprint(root, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			got = strings.TrimSpace(got)
			if want := strings.Join(tt.want, "\n"); got != want {
				t.Errorf("invalid output:\ngot:\n%v\n\nwant:\n%s", got, want)
			}
		})
	}
}
