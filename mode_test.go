package cpio_test

import (
	"io/fs"
	"strconv"
	"testing"

	"github.com/korylprince/go-cpio-odc"
)

var modeTests = []fs.FileMode{
	0755 | fs.ModeDir,
	0644 | fs.ModeSymlink,
	0640 | fs.ModeDevice,
	0604 | fs.ModeNamedPipe,
	0511 | fs.ModeSocket,
	0600 | fs.ModeCharDevice,
	0644 | fs.ModeSticky | fs.ModeSetuid | fs.ModeSetgid,
}

// tests from running cpio -c -o on various files
var modeMarshalTests = map[string]fs.FileMode{
	"100644": 0644,
	"040755": 0755 | fs.ModeDir,
	"120777": 0777 | fs.ModeSymlink,
	"060660": 0660 | fs.ModeDevice,
	"010600": 0600 | fs.ModeNamedPipe,
	"140660": 0660 | fs.ModeSocket,
	"020666": 0666 | fs.ModeCharDevice,
	"041755": 0755 | fs.ModeDir | fs.ModeSticky,
	"106755": 0755 | fs.ModeSetuid | fs.ModeSetgid,
}

func TestMode(t *testing.T) {
	t.Parallel()

	for _, want := range modeTests {
		i := cpio.MarshalFileMode(want)
		have := cpio.UnmarshalFileMode(i)
		if have != want {
			t.Errorf(`expected file modes to be equal: want: %o, have: %o`, want, have)
		}
	}

	for test, want := range modeMarshalTests {
		mode, _ := strconv.ParseUint(test, 8, 64)
		have := cpio.UnmarshalFileMode(mode)
		if have != want {
			t.Errorf(`expected file modes to be equal: want: test: %s, want: %o, have: %o`, test, want, have)
		}
	}
}
