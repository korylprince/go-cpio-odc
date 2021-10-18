[![pkg.go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/korylprince/go-cpio-odc)

# About

`go-cpio-odc` is a pure Go reader and writer for POSIX.1 portable format cpio archives (also known as odc).

# Installing

`go get github.com/korylprince/go-cpio-odc/v3`

If you have any issues or questions [create an issue](https://github.com/korylprince/go-cpio-odc/issues).


# Usage

The library provides a `Reader` and `Writer`, as well as an `fs.FS` to browse the archive and a `Writer.WriteFS` method to write and entire `fs.FS` to the archive.

See examples on [pkg.go.dev](https://pkg.go.dev/github.com/korylprince/go-cpio-odc/#pkg-examples).

# Testing

`go test -v`

Some tests require GNU's `cpio` to be in `$PATH`. If it's not available, those test will be skipped.
