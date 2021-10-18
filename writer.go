package cpio

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
)

type Writer struct {
	w  io.Writer
	bs int
	n  int64
}

// NewWriter returns a new Writer using the given block size. If blockSize is 0, DefaultBlockSize will be used.
// Note: the writer doesn't actually write in blockSize blocks. Instead the output is padded with NULL bytes to make the total size a multiple of blockSize.
func NewWriter(w io.Writer, blockSize int) *Writer {
	if blockSize == 0 {
		blockSize = DefaultBlockSize
	}

	return &Writer{w: w, bs: blockSize, n: 0}
}

func toOctal(n interface{}, size int) []byte {
	s := fmt.Sprintf(fmt.Sprintf("%%0%do", size), n)
	return []byte(s[len(s)-size:])
}

// WriteFile writes f to the Writer
func (w *Writer) WriteFile(f *File) error {
	buf := make([]byte, 0, headerSize)

	buf = append(buf, HeaderMagic...)
	buf = append(buf, toOctal(f.Device, fieldSize)...)
	buf = append(buf, toOctal(f.Inode, fieldSize)...)
	buf = append(buf, toOctal(MarshalFileMode(f.FileMode), fieldSize)...)
	buf = append(buf, toOctal(f.UID, fieldSize)...)
	buf = append(buf, toOctal(f.GID, fieldSize)...)
	buf = append(buf, toOctal(f.NLink, fieldSize)...)
	buf = append(buf, toOctal(f.RDev, fieldSize)...)
	var mtime int64
	if !f.ModifiedTime.IsZero() {
		mtime = f.ModifiedTime.Unix()
	}
	buf = append(buf, toOctal(mtime, largeFieldSize)...)
	buf = append(buf, toOctal(len(f.Path)+1, fieldSize)...)
	buf = append(buf, toOctal(len(f.Body), largeFieldSize)...)

	n, err := w.w.Write(buf)
	if err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}
	w.n += int64(n)

	n, err = w.w.Write([]byte(f.Path + "\x00"))
	if err != nil {
		return fmt.Errorf("could not write path: %w", err)
	}
	w.n += int64(n)

	n, err = w.w.Write(f.Body)
	if err != nil {
		return fmt.Errorf("could not write body: %w", err)
	}
	w.n += int64(n)

	return nil
}

// WriteFS runs fs.WalkDir on f and writes every file and directory encountered.
// If skipErrs is false, any error encountered while walking the FS will halt the process and the error will be returned
func (w *Writer) WriteFS(f fs.FS, skipErrs bool) error {
	if err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		file, err := f.Open(path)
		if err != nil {
			if skipErrs {
				return nil
			}
			return fmt.Errorf("could not open %s: %w", path, err)
		}

		fi, err := file.Stat()
		if err != nil {
			if skipErrs {
				return nil
			}
			return fmt.Errorf("could not stat %s: %w", path, err)
		}

		cFile := statFile(fi)
		cFile.Path = path

		if fi.Mode().IsRegular() {
			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(file); err != nil {
				if skipErrs {
					return nil
				}
				return fmt.Errorf("could not read %s: %w", path, err)
			}
			cFile.Body = buf.Bytes()
			cFile.buf = buf
		}

		if err = w.WriteFile(cFile); err != nil {
			return fmt.Errorf("could not write %s: %w", path, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error encountered while walking FS: %w", err)
	}
	return nil
}

// Close writes out the trailer file, pads the writer to a multiple of blockSize, and returns the total bytes written
func (w *Writer) Close() (int64, error) {
	n, err := w.w.Write(trailer)
	if err != nil {
		return w.n, fmt.Errorf("could not write trailer file: %w", err)
	}
	w.n += int64(n)

	if b := w.n % int64(w.bs); b > 0 {
		n, err := w.w.Write(make([]byte, w.bs-int(b)))
		if err != nil {
			return w.n, fmt.Errorf("could not write block padding: %w", err)
		}
		w.n += int64(n)
	}

	return w.n, nil
}
