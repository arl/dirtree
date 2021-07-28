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
			name: "ModeAll",
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
			name: "ExcludeRoot",
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
			name: "IncludeRoot(true)",
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
				t.Errorf("got:\n%v\nwant:\n%s", got, want)
			}
		})
	}
}
