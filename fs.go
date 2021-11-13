package cpio

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"time"
)

func (f *File) Name() string {
	return path.Base(f.Path)
}

func (f *File) Size() int64 {
	return int64(len(f.Body))
}

func (f *File) Mode() fs.FileMode {
	return f.FileMode
}

func (f *File) ModTime() time.Time {
	return f.ModifiedTime
}

func (f *File) IsDir() bool {
	return f.FileMode.IsDir()
}

func (f *File) Sys() interface{} {
	return f
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *File) Read(b []byte) (int, error) {
	return f.buf.Read(b)
}

func (f *File) Close() error {
	return nil
}

func (f *File) Type() fs.FileMode {
	return f.FileMode
}

func (f *File) Info() (fs.FileInfo, error) {
	return f, nil
}

type DirFile struct {
	*File
	entries  []fs.DirEntry
	entryidx int
}

// ReadDir implements the ReadDirFile interface
func (f *DirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	clean := path.Clean(f.Path)
	if len(clean) > 0 && clean[0] == '/' {
		clean = clean[1:]
	}
	if f.entries == nil {
		var entries []fs.DirEntry
		for p, file := range f.fs.files {
			if clean != p && path.Dir(p) == clean {
				entries = append(entries, file.(fs.DirEntry))
			}
		}
		f.entries = entries
	}

	if f.entryidx <= len(f.entries) {
		if n <= 0 || f.entryidx+n >= len(f.entries) {
			idx := f.entryidx
			f.entryidx = len(f.entries)
			if n <= 0 {
				return f.entries[idx:], nil
			}
			return f.entries[idx:], io.EOF
		}
		idx := f.entryidx
		f.entryidx += n
		return f.entries[idx : idx+n], nil
	}

	return nil, io.EOF
}

type FS struct {
	files map[string]fs.File
}

// NewFS returns an fs.FS by parsing a cpio archive from the given reader.
// Note: the entire reader is buffered in memory, since cpio archives don't contain a table of contents
func NewFS(r io.Reader) (fs.FS, error) {
	cfs := &FS{files: make(map[string]fs.File)}

	rdr := NewReader(r)
	var (
		f   *File
		err error
	)

	dirs := make(map[string]struct{})

	for f, err = rdr.Next(); err == nil; f, err = rdr.Next() {
		clean := path.Clean(f.Path)
		if len(clean) > 0 && clean[0] == '/' {
			clean = clean[1:]
		}
		for dir := path.Dir(clean); dir != "."; dir = path.Dir(dir) {
			dirs[dir] = struct{}{}
		}
		if f.IsDir() {
			cfs.files[clean] = &DirFile{File: f}
		} else {
			cfs.files[clean] = f
		}
		f.fs = cfs
	}
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("could not parse as cpio: %w", err)
		}
	}

	for d := range dirs {
		if _, ok := cfs.files[d]; ok {
			continue
		}
		cfs.files[d] = &DirFile{File: &File{Path: d, FileMode: 0755 | fs.ModeDir, fs: cfs}}
	}
	// root directory
	cfs.files["."] = &DirFile{File: &File{Path: ".", FileMode: 0755 | fs.ModeDir, fs: cfs}}

	return cfs, nil
}

// Open implements the fs.FS interface
func (f *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	file, ok := f.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return file, nil
}

// ReadFile implements the fs.ReadFileFS interface
func (f *FS) ReadFile(name string) ([]byte, error) {
	file, err := f.Open(name)
	if err != nil {
		return nil, err
	}

	return file.(*File).buf.Bytes(), nil
}
