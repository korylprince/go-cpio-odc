package cpio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

func read(r io.Reader, buf []byte, size int) error {
	if n, err := r.Read(buf); err != nil {
		return fmt.Errorf("could not read %d bytes: %w", size, err)
	} else if n != size {
		return fmt.Errorf("could not read: want: %d, have: %d", size, n)
	}
	return nil
}

type Reader struct {
	r    io.Reader
	buf  []byte
	lbuf []byte
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:    r,
		buf:  make([]byte, fieldSize),
		lbuf: make([]byte, largeFieldSize),
	}
}

// Next parses and returns the next file in the archive. If no files are left in the archive, err will be io.EOF
func (r *Reader) Next() (*File, error) {
	var err error

	// read magic
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read magic: %w", err)
	}
	if !bytes.Equal(r.buf, HeaderMagic) {
		return nil, fmt.Errorf("invalid magic: %s", r.buf)
	}

	f := &File{}

	// read device
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read device: %w", err)
	}
	f.Device, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse device: %w", err)
	}

	// read inode
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read inode: %w", err)
	}
	f.Inode, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse inode: %w", err)
	}

	// read mode
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read mode: %w", err)
	}
	mode, err := strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse mode: %w", err)
	}
	f.FileMode = UnmarshalFileMode(mode)

	// read uid
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read uid: %w", err)
	}
	f.UID, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse uid: %w", err)
	}

	// read gid
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read gid: %w", err)
	}
	f.GID, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse gid: %w", err)
	}

	// read link number
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read link number: %w", err)
	}
	f.NLink, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse link number: %w", err)
	}

	// read device number
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read device number: %w", err)
	}
	f.RDev, err = strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse device number: %w", err)
	}

	// read modified time
	if err := read(r.r, r.lbuf, largeFieldSize); err != nil {
		return nil, fmt.Errorf("could not read modified time: %w", err)
	}
	modTime, err := strconv.ParseUint(string(r.lbuf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse modified time: %w", err)
	}
	f.ModifiedTime = time.Unix(int64(modTime), 0)

	// read path length
	if err := read(r.r, r.buf, fieldSize); err != nil {
		return nil, fmt.Errorf("could not read path length: %w", err)
	}
	plen, err := strconv.ParseUint(string(r.buf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse path length: %w", err)
	}

	// read body length
	if err := read(r.r, r.lbuf, largeFieldSize); err != nil {
		return nil, fmt.Errorf("could not read body length: %w", err)
	}
	blen, err := strconv.ParseUint(string(r.lbuf), octalBits, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse body length: %w", err)
	}

	// read path
	pbuf := make([]byte, plen)
	if err := read(r.r, pbuf, int(plen)); err != nil {
		return nil, fmt.Errorf("could not read path: %w", err)
	}
	if pbuf[plen-1] != '\x00' {
		return nil, fmt.Errorf("could not read path: invalid terminator: want: %#x, have: %#x", '\x00', pbuf[plen-1])
	}

	// check for trailer
	if bytes.Equal(pbuf[:plen-1], TrailerPath) {
		return nil, io.EOF
	}
	f.Path = string(pbuf[:plen-1])

	// read body
	bbuf := make([]byte, blen)
	if err := read(r.r, bbuf, int(blen)); err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}
	f.Body = bbuf
	f.buf = bytes.NewBuffer(f.Body)

	return f, nil
}

// ReadFile returns a *File for the given path. Where possible (on unix), it will use syscalls to get fields not available to os.Stat
func ReadFile(path string) (*File, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not stat file: %w", err)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	file := statFile(fi)
	file.Path = path
	file.Body = buf
	file.buf = bytes.NewBuffer(buf)
	return file, nil
}
