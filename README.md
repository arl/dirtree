[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/arl/dirtree)
[![Test Actions Status](https://github.com/arl/dirtree/workflows/Test/badge.svg)](https://github.com/arl/dirtree/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/arl/dirtree)](https://goreportcard.com/report/github.com/arl/dirtree)
[![codecov](https://codecov.io/gh/arl/dirtree/branch/main/graph/badge.svg)](https://codecov.io/gh/arl/dirtree)


Dirtree
========

Dirtree recursively walks a directory structure and prints one line per file,
plus additional information such as size, permissions, CRC-32 hash of files
content, etc. Its output makes it super handy to be used as golden file, when
testing functions which outcome is to create a directory structure with some
files in it.

```go
ls, err := dirtree.Print(root, dirtree.ModeAll)
if err != nil {
    log.Fatalf("dirtree error: %v", err)
}
// output: 
// d 775 sym=0            crc=n/a      .
// d 775 sym=0            crc=n/a      A
// d 775 sym=0            crc=n/a      A/B
// ? 777 sym=1            crc=n/a      A/B/symdirA
// f 775 sym=0 13b        crc=0451ac5e A/file1
// ? 777 sym=1            crc=n/a      A/symfile1
```


`dirtree` command-line tool
---------------------------

```sh
go install github.com/arl/dirtree/cmd/dirtree@latest
```

Installs `dirtree`, a super basic CLI wrapper over the dirtree Go module. 

TODO
----

  - extend API
    +  exclude files
    +  exclude root
    +  limit depth


License
-------

- [MIT License](LICENSE)
