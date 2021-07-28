[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/arl/dirtree)
[![Test Actions Status](https://github.com/arl/dirtree/workflows/Test/badge.svg)](https://github.com/arl/dirtree/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/arl/dirtree)](https://goreportcard.com/report/github.com/arl/dirtree)
[![codecov](https://codecov.io/gh/arl/dirtree/branch/main/graph/badge.svg)](https://codecov.io/gh/arl/dirtree)


Dirtree
========

Dirtree recursively walks a directory structure and prints one line per file,
plus additional information such as directory, file size and a CRC-32 hash of
its content for quick comparison. This is mostly useful for quickly checking the
differences between 2 directory structures, in tests for example.  

```go
ls, err := dirtree.Print(root, dirtree.ModeAll)
if err != nil {
    log.Fatalf("dirtree error: %v", err)
}
// output: 
// d            crc=n/a      .
// d            crc=n/a      A
// d            crc=n/a      A/B
// ?            crc=n/a      A/B/symdirA
// f 13b        crc=0451ac5e A/file1
// ?            crc=n/a      A/symfile1
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
