// package cpio is a library to read and write POSIX.1 (also known as odc or old portable ASCII) cpio archive files
package cpio

import (
	"bytes"
	"io/fs"
	"time"
)

var HeaderMagic = []byte("070707")
var TrailerPath = []byte("TRAILER!!!")
var trailer = []byte("0707070000000000000000000000000000000000010000000000000000000001300000000000TRAILER!!!\x00")

const octalBits = 8
const fieldSize = 6
const largeFieldSize = 11

// DefaultBlockSize is the default block size when writing. The default matches GNU cpio
const DefaultBlockSize = 512

const headerSize = 9*6 + 2*11

// File is a file stored in a cpio archive. All uint64 fields are limited to 48 bits, and any higher bits are truncated by Writer
type File struct {
	Device       uint64
	Inode        uint64
	FileMode     fs.FileMode
	UID          uint64
	GID          uint64
	NLink        uint64
	RDev         uint64
	ModifiedTime time.Time
	Path         string
	Body         []byte
	buf          *bytes.Buffer
	fs           *FS
}
