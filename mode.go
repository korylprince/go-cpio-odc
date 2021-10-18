package cpio

import "io/fs"

const (
	MaskFileType   uint64 = 0170000
	ModeSocket     uint64 = 0140000
	ModeSymlink    uint64 = 0120000
	ModeRegular    uint64 = 0100000
	ModeDevice     uint64 = 0060000
	ModeDir        uint64 = 0040000
	ModeCharDevice uint64 = 0020000
	ModeNamedPipe  uint64 = 0010000
	ModeSetuid     uint64 = 0004000
	ModeSetgid     uint64 = 0002000
	ModeSticky     uint64 = 0001000
)

func UnmarshalFileMode(m uint64) fs.FileMode {
	mode := fs.FileMode(m).Perm()

	if m&ModeSetuid != 0 {
		mode |= fs.ModeSetuid
	}
	if m&ModeSetgid != 0 {
		mode |= fs.ModeSetgid
	}
	if m&ModeSticky != 0 {
		mode |= fs.ModeSticky
	}

	m &= MaskFileType
	if m == ModeSocket {
		mode |= fs.ModeSocket
	}
	if m == ModeSymlink {
		mode |= fs.ModeSymlink
	}
	if m == ModeDevice {
		mode |= fs.ModeDevice
	}
	if m == ModeDir {
		mode |= fs.ModeDir
	}
	if m == ModeCharDevice {
		mode |= fs.ModeCharDevice
	}
	if m == ModeNamedPipe {
		mode |= fs.ModeNamedPipe
	}

	return mode
}

func MarshalFileMode(m fs.FileMode) uint64 {
	mode := uint64(m.Perm())

	if m&fs.ModeSetuid != 0 {
		mode |= ModeSetuid
	}
	if m&fs.ModeSetgid != 0 {
		mode |= ModeSetgid
	}
	if m&fs.ModeSticky != 0 {
		mode |= ModeSticky
	}

	m &= fs.ModeType
	if m == 0 {
		mode |= ModeRegular
		return mode
	}

	if m == fs.ModeSocket {
		mode |= ModeSocket
	}
	if m == fs.ModeSymlink {
		mode |= ModeSymlink
	}
	if m == fs.ModeDevice {
		mode |= ModeDevice
	}
	if m == fs.ModeDir {
		mode |= ModeDir
	}
	if m == fs.ModeCharDevice {
		mode |= ModeCharDevice
	}
	if m == fs.ModeNamedPipe {
		mode |= ModeNamedPipe
	}

	return mode
}
