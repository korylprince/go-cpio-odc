package cpio_test

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"

	"github.com/korylprince/go-cpio-odc"
)

func ExampleReader() {
	f, err := os.Open("testdata/test.cpio")
	if err != nil {
		log.Fatalln("could not open archive:", err)
	}

	r := cpio.NewReader(f)
	var (
		files []*cpio.File
		file  *cpio.File
	)
	for file, err = r.Next(); err == nil; file, err = r.Next() {
		files = append(files, file)
	}
	if errors.Is(err, io.EOF) {
		log.Println("found", len(files), "files")
		return
	}
	if err != nil {
		log.Fatalln("could not read archive:", err)
	}
}

func ExampleFS() {
	f, err := os.Open("testdata/test.cpio")
	if err != nil {
		log.Fatalln("could not open archive:", err)
	}

	fs, err := cpio.NewFS(f)
	if err != nil {
		log.Fatalln("could not create fs:", err)
	}

	file, err := fs.Open("test.txt")
	if err != nil {
		log.Fatalln("could not open file:", err)
	}

	info, err := file.Stat()
	if err != nil {
		log.Fatalln("could not stat file:", err)
	}

	log.Println("test.txt mode:", info.Mode())
}

func ExampleWriter() {
	buf := new(bytes.Buffer)
	w := cpio.NewWriter(buf, 0)
	f := &cpio.File{
		FileMode: 0644,
		Path:     "hello.txt",
		Body:     []byte("Hello, world!\n"),
	}

	if err := w.WriteFile(f); err != nil {
		log.Fatalln("could not write file:", err)
	}

	n, err := w.Close()
	if err != nil {
		log.Fatalln("could not close file:", err)
	}

	log.Println("wrote", n, "bytes")
}

func ExampleWriter_WriteFS() {
	buf := new(bytes.Buffer)
	w := cpio.NewWriter(buf, 0)
	dir := os.DirFS("testdata")
	if err := w.WriteFS(dir, false); err != nil {
		log.Fatalln("could not write fs:", err)
	}

	n, err := w.Close()
	if err != nil {
		log.Fatalln("could not close file:", err)
	}

	log.Println("wrote", n, "bytes")
}
