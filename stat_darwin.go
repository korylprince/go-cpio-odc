package cpio

import (
	"syscall"
	"time"
)

func statToFile(info *syscall.Stat_t) *File {
	return &File{
		Device: uint64(info.Dev), Inode: info.Ino,
		UID: uint64(info.Uid), GID: uint64(info.Gid),
		NLink: uint64(info.Nlink), RDev: uint64(info.Rdev),
		ModifiedTime: time.Unix(info.Mtimespec.Sec, info.Mtimespec.Nsec),
	}
}
