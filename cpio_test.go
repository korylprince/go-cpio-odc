package cpio_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/korylprince/go-cpio-odc"
)

func TestCPIO(t *testing.T) {
	t.Parallel()

	buf, err := os.ReadFile("testdata/test.cpio")
	if err != nil {
		t.Fatalf("could not read testdata: %v", err)
	}

	r := cpio.NewReader(bytes.NewBuffer(buf))
	var (
		files []*cpio.File
		file  *cpio.File
	)
	for file, err = r.Next(); err == nil; file, err = r.Next() {
		files = append(files, file)
	}
	if err == nil {
		t.Errorf("expected archive error to be EOF: want: %v, have: %v", io.EOF, err)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("could not read archive: %v", err)
	}

	have := new(bytes.Buffer)
	w := cpio.NewWriter(have, 0)
	for _, f := range files {
		if err = w.WriteFile(f); err != nil {
			t.Fatalf(`could not write file "%s": %v`, f.Path, err)
		}
	}

	n, err := w.Close()
	if err != nil {
		t.Fatalf("could not close archive: %v", err)
	}

	if have.Len() != int(n) || n%cpio.DefaultBlockSize != 0 {
		t.Errorf("expected archive blocks to be correct: want: %v, have: %v", n*cpio.DefaultBlockSize, have.Len())
	}

	if !bytes.Equal(have.Bytes(), buf) {
		t.Error("expected output to be equal to testdata")
	}
}

func gnuCPIO() (string, error) {
	// check for cpio command
	cpiocmd, err := exec.LookPath("cpio")
	if err != nil {
		return "", fmt.Errorf("could not find cpio executable: %w", err)
	}

	output, err := exec.Command(cpiocmd, "--version").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not get cpio version: %w", err)
	}

	if !strings.Contains(string(output), "GNU cpio") {
		return "", errors.New("cpio is not GNU version")
	}

	return cpiocmd, nil
}

func TestReadFile(t *testing.T) {
	t.Parallel()

	cpiopath, err := gnuCPIO()
	if err != nil {
		t.Skip(err)
	}

	// generate our version
	f, err := cpio.ReadFile("testdata/test.txt")
	if err != nil {
		t.Fatalf("could not read testdata: %v", err)
	}

	buf := new(bytes.Buffer)
	w := cpio.NewWriter(buf, 0)
	if err = w.WriteFile(f); err != nil {
		t.Fatalf("could not write file: %v", err)
	}
	if _, err = w.Close(); err != nil {
		t.Fatalf("could not close archive: %v", err)
	}
	have := buf.Bytes()

	// generate their version
	buf = new(bytes.Buffer)
	cmd := exec.Command(cpiopath, "-c", "-o")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("could not get stdin: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("could not get stdin: %v", err)
	}

	if err = cmd.Start(); err != nil {
		t.Fatalf("could not start cpio: %v", err)
	}
	if _, err = stdin.Write([]byte("testdata/test.txt")); err != nil {
		t.Fatalf("could not write to stdin: %v", err)
	}
	if err = stdin.Close(); err != nil {
		t.Fatalf("could not close stdin: %v", err)
	}

	if _, err = buf.ReadFrom(stdout); err != nil {
		t.Fatalf("could not read from stdout: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		t.Fatalf("cpio did not exit successfully: %v", err)
	}
	want := buf.Bytes()

	// check if they're the same
	if !bytes.Equal(have, want) {
		t.Error("library output did not match cpio")
	}
}

func TestWriteFS(t *testing.T) {
	t.Parallel()

	cpiopath, err := gnuCPIO()
	if err != nil {
		t.Skip(err)
	}

	// generate our version
	dir := os.DirFS("testdata")

	buf := new(bytes.Buffer)
	w := cpio.NewWriter(buf, 0)
	if err = w.WriteFS(dir, false); err != nil {
		t.Fatalf("could not write FS: %v", err)
	}
	if _, err = w.Close(); err != nil {
		t.Fatalf("could not close archive: %v", err)
	}
	have := buf.Bytes()

	// generate their version
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working directory: %v", err)
	}
	buf = new(bytes.Buffer)
	cmd := exec.Command(cpiopath, "-c", "-o")
	cmd.Dir = path.Join(cwd, "testdata")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("could not get stdin: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("could not get stdin: %v", err)
	}

	if err = cmd.Start(); err != nil {
		t.Fatalf("could not start cpio: %v", err)
	}
	if _, err = stdin.Write([]byte(".\ntest.cpio\ntest.txt")); err != nil {
		t.Fatalf("could not write to stdin: %v", err)
	}
	if err = stdin.Close(); err != nil {
		t.Fatalf("could not close stdin: %v", err)
	}

	if _, err = buf.ReadFrom(stdout); err != nil {
		t.Fatalf("could not read from stdout: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		t.Fatalf("cpio did not exit successfully: %v", err)
	}
	want := buf.Bytes()

	// check if they're the same
	if !bytes.Equal(have, want) {
		t.Error("library output did not match cpio")
	}
}

func TestFS(t *testing.T) {
	t.Parallel()

	orig, err := os.ReadFile("testdata/test.cpio")
	if err != nil {
		t.Fatalf("could not read testdata: %v", err)
	}

	fs, err := cpio.NewFS(bytes.NewBuffer(orig))
	if err != nil {
		t.Fatalf("could not create FS: %v", err)
	}

	// get files from original
	r := cpio.NewReader(bytes.NewBuffer(orig))
	wantFiles := make(map[string]*cpio.File)
	var file *cpio.File
	for file, err = r.Next(); err == nil; file, err = r.Next() {
		wantFiles[file.Path] = file
	}
	if err == nil {
		t.Errorf("expected original archive error to be EOF: want: %v, have: %v", io.EOF, err)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("could not read original archive: %v", err)
	}

	// get files from WriteFS
	buf := new(bytes.Buffer)
	w := cpio.NewWriter(buf, 0)
	if err = w.WriteFS(fs, false); err != nil {
		t.Fatalf("could not write fs: %v", err)
	}
	if _, err := w.Close(); err != nil {
		t.Fatalf("could not close fs archive: %v", err)
	}

	r = cpio.NewReader(bytes.NewBuffer(buf.Bytes()))
	haveFiles := make(map[string]*cpio.File)
	for file, err = r.Next(); err == nil; file, err = r.Next() {
		haveFiles[file.Path] = file
	}
	if err == nil {
		t.Errorf("expected fs archive error to be EOF: want: %v, have: %v", io.EOF, err)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("could not read fs archive: %v", err)
	}

	for fp, file1 := range wantFiles {
		clean := path.Clean(fp)
		if len(clean) > 0 && clean[0] == '/' {
			clean = clean[1:]
		}
		if file2, ok := haveFiles[clean]; !ok {
			t.Errorf("expected to find %s", fp)
		} else {
			file1.Path = clean
			if !reflect.DeepEqual(file1, file2) {
				t.Errorf("expected %s to be equal:", fp)
			}
		}
	}
}
