
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/arl/dirtree)
[![Go Report Card](https://goreportcard.com/badge/github.com/arl/dirtree)](https://goreportcard.com/report/github.com/arl/dirtree)
[![codecov](https://codecov.io/gh/arl/dirtree/branch/main/graph/badge.svg)](https://codecov.io/gh/arl/dirtree)

[![Tests-linux actions Status](https://github.com/arl/dirtree/actions/workflows/tests-linux.yml/badge.svg)](https://github.com/arl/dirtree/actions)
[![Tests-others Actions Status](https://github.com/arl/dirtree/actions/workflows/tests-others.yml/badge.svg)](https://github.com/arl/dirtree/actions)

Dirtree
========

**Dirtree** lists the files in a directory, or an [fs.FS](https://pkg.go.dev/io/fs#FS), in a deterministic, configurable and cross-platform manner.

Very useful for testing, golden files, etc.


Usage
-----

Add the `dirtree` module as dependency:

    go get github.com/arl/dirtree@latest


 - Basic usage

To list the `./testdata` directory, recursively, to standard output:

```go
err := dirtree.Write(os.Stdout, "./testdata")
if err != nil {
	log.Fatal(err)
}
```

Using options, you can control what files are listed and what information is shown (size, CRC32, type, etc.).


# Documentation

## Walk your OS filesystem or an [`fs.FS`](https://pkg.go.dev/io/fs#FS)?

Use `dirtree.Write` and `dirtree.Sprint` to walk your filesystem, or,  
use `dirtree.WriteFS` and `dirtree.SprintFS` to walk an [fs.FS](https://pkg.go.dev/io/fs#FS).  
These functions are useful when you're interested by the whole directory
listing, to use for golden tests or displaying to user.

If you don't need to print the directory tree, you can use `dirtree.List`, it
returns a slice of `dirtree.Entry` which you can examine programmaticaly.

All above functions accept a variable number (possibly none) of options.
For example:

```go
dirtree.Write(os.Stdout, "dir", dirtree.Depth(2), dirtree.ModeSize | dirtree.ModeCRC32)
```

In the following examples the root directory has the following content:
```
.
├── bar
│   ├── dir1
│   │   └── passwords
│   └── dir2
├── baz
│   └── a
│       └── b
│           └── c
│               └── nested
├── foo
│   ├── dir1
│   └── dir2
│       └── secrets
├── other-stuff.mp3
└── symlink -> foo/dir2/secrets
```


### Default output (no options)

Calling `dirtree.Write` without any option will show, as in:
```go
dirtree.Write(os.Stdout, "dir")
```
shows:
```
d            .
d            bar
d            bar/dir1
f 0b         bar/dir1/passwords
d            bar/dir2
d            baz
d            baz/a
d            baz/a/b
d            baz/a/b/c
f 1407216b   baz/a/b/c/nested
d            foo
d            foo/dir1
d            foo/dir2
f 7922820b   foo/dir2/secrets
f 39166b     other-stuff.mp3
?            symlink
```


### `Type` option

The `dirtree.Type` option limits the files present in the listing based on their types.
It's a string that may contain one or more characters:
  - `f` for regular files
  - `d` for directories
  - `?` for anything else (symlink, etc.)

For example, `dirtree.Type("f")` will only show regular files while
`dirtree.Type("fd")` will both show regular files and directories.

By default, `dirtree` shows all types if the `Type` option is not provided.

```go
dirtree.Write(os.Stdout, "dir", dirtree.Type("f?"))
```

displays:

```
f 0b         crc=00000000 bar/dir1/passwords
f 1407216b   crc=733eee4d baz/a/b/c/nested
f 7922820b   crc=fe02449a foo/dir2/secrets
f 39166b     crc=d298754e other-stuff.mp3
?            crc=n/a      symlink
```



### `PrintMode` option

The `dirtree.PrintMode` option is a bitset controlling the amount of information
to show for each listed file.

   - `dirtree.ModeType` prints 'd', 'f' or '?', depending on the file type,
     directory, regular file or anything else.
   - `dirtree.ModeSize` shows the file size in bytes, for regular files only.
   - `dirtree.ModeCRC32` shows a CRC32 checksum, for regular files only.


`dirtree.ModeDefault` combines `dirtree.ModeType` | `dirtree.ModeSize` and
`dirtree.ModeAll` shows all information about all files:


```go
dirtree.Write(os.Stdout, "dir", dirtree.ModeAll)
```
displays:

```
d            crc=n/a      .
d            crc=n/a      bar
d            crc=n/a      bar/dir1
f 0b         crc=00000000 bar/dir1/passwords
d            crc=n/a      bar/dir2
d            crc=n/a      baz
d            crc=n/a      baz/a
d            crc=n/a      baz/a/b
d            crc=n/a      baz/a/b/c
f 1407216b   crc=733eee4d baz/a/b/c/nested
d            crc=n/a      foo
d            crc=n/a      foo/dir1
d            crc=n/a      foo/dir2
f 7922820b   crc=fe02449a foo/dir2/secrets
f 39166b     crc=d298754e other-stuff.mp3
?            crc=n/a      symlink
```

### `Ignore` files

The `dirtree.Ignore` option allows to ignore files matching a pattern. The path
relative to the chosen root is matched against the pattern. Ignore follows the
syntax used and described with the `filepath.Match` function. Before checking if
it matches a pattern, a path is first converted to its slash ('/') based
version, to ensure cross-platform consistency of the dirtree package.

`dirtree.Ignore` can be provided multiple times to ignore multiple patterns. A
file is ignored from the listing as long as at it matches at least one
`dirtree.Ignore` pattern. Also, `dirtree.Ignore` has precedence over
`dirtree.Match`.

The pattern syntax is that of [filepath.Match](https://pkg.go.dev/path/filepath#Match):
```
pattern:
	{ term }
term:
	'*'         matches any sequence of non-Separator characters
	'?'         matches any single non-Separator character
	'[' [ '^' ] { character-range } ']'
	            character class (must be non-empty)
	c           matches character c (c != '*', '?', '\\', '[')
	'\\' c      matches character c

character-range:
	c           matches character c (c != '\\', '-', ']')
	'\\' c      matches character c
	lo '-' hi   matches character c for lo <= c <= hi
```

```go
dirtree.Write(os.Stdout, dir, dirtree.Ignore("*/dir1"))
```
prints:
```
d            .
d            bar
f 0b         bar/dir1/passwords
d            bar/dir2
d            baz
d            baz/a
d            baz/a/b
d            baz/a/b/c
f 1407216b   baz/a/b/c/nested
d            foo
d            foo/dir2
f 7922820b   foo/dir2/secrets
f 39166b     other-stuff.mp3
?            symlink
```


### `Match` to limit the listing to files matching a pattern

The `dirtree.Match` option limits the listing to files that match a pattern. The
path relative to the chosen root is matched against the pattern. `dirtree.Match`
follows the syntax used and described with the `filepath.Match` function. Before
checking if it matches a pattern, a path is first converted to its slash ('/')
based version, to ensure cross-platform consistency of the dirtree package.

`dirtree.Match` can be provided multiple times to match multiple patterns. A
file is included in the listing as long as at it matches at least one
`dirtree.Match` pattern, unless it matches an `dirtree.Ignore` pattern (since
`dirtree.Ignore` has precedence over `dirtree.Match`).

The pattern syntax is that of [filepath.Match](https://pkg.go.dev/path/filepath#Match).

```go
dirtree.Write(os.Stdout, dir, dirtree.Match("*/*[123]"), dirtree.Match("*[123]"))
```
prints:
```
d            bar/dir1
d            bar/dir2
d            foo/dir1
d            foo/dir2
f 39166b     other-stuff.mp3
```


### Directory `Depth`

The `dirtree.Depth` option is an integer that controls the maximum depth to
descend into, starting from the root directory. Everything below that depth
won't be shown.


```go
dirtree.Write(os.Stdout, dir, dirtree.Depth(1))
```
reports:
```
d            .
d            bar
d            baz
d            foo
f 39166b     other-stuff.mp3
?            symlink
```

### `ExcludeRoot`

`dirtree.ExcludeRoot` hides the root directory in the listing.
```go
dirtree.Write(os.Stdout, dir, dirtree.ExcludeRoot)
```

```
d            bar
d            bar/dir1
f 0b         bar/dir1/passwords
d            bar/dir2
d            baz
d            baz/a
d            baz/a/b
d            baz/a/b/c
f 1407216b   baz/a/b/c/nested
d            foo
d            foo/dir1
d            foo/dir2
f 7922820b   foo/dir2/secrets
f 39166b     other-stuff.mp3
?            symlink
```




## TODO
 - streaming API (for large number of files)

License
-------

- [MIT License](LICENSE)
