//go:build !darwin
// +build !darwin

package cpio

import (
	"syscall"
	"time"
)

func statToFile(info *syscall.Stat_t) *File {
	return &File{
		Device: info.Dev, Inode: info.Ino,
		UID: uint64(info.Uid), GID: uint64(info.Gid),
		NLink: info.Nlink, RDev: info.Rdev,
		ModifiedTime: time.Unix(info.Mtim.Sec, info.Mtim.Nsec),
	}
}
