//go:build !darwin && !windows
// +build !darwin,!windows

package cpio

import (
	"io/fs"
	"syscall"
	"time"
)

// statFile builds a *File with the information available in fi. The caller is responsible for filling in the body
func statFile(fi fs.FileInfo) *File {
	switch info := fi.Sys().(type) {
	// on unix
	case *syscall.Stat_t:
		return &File{
			Device: uint64(info.Dev), Inode: info.Ino,
			FileMode: fi.Mode(),
			UID:      uint64(info.Uid), GID: uint64(info.Gid),
			NLink: uint64(info.Nlink), RDev: uint64(info.Rdev),
			ModifiedTime: time.Unix(int64(info.Mtim.Sec), int64(info.Mtim.Nsec)),
		}

	// cpioception
	case *File:
		return &File{
			Device: info.Device, Inode: info.Inode,
			FileMode: info.FileMode,
			UID:      info.UID, GID: info.GID,
			NLink: info.NLink, RDev: info.RDev,
			ModifiedTime: info.ModifiedTime,
		}

	case *DirFile:
		return &File{
			Device: info.Device, Inode: info.Inode,
			FileMode: info.FileMode,
			UID:      info.UID, GID: info.GID,
			NLink: info.NLink, RDev: info.RDev,
			ModifiedTime: info.ModifiedTime,
		}
	}

	// on generic fs.FS, see: https://www.mkssoftware.com/docs/man5/stat.5.asp
	return &File{
		Device: 0, Inode: 3,
		FileMode: fi.Mode(),
		UID:      0, GID: 0,
		NLink: 1, RDev: 0,
		ModifiedTime: fi.ModTime(),
	}
}
