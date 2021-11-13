package cpio

import (
	"io/fs"
)

// statFile builds a *File with the information available in fi. The caller is responsible for filling in the body
func statFile(fi fs.FileInfo) *File {
	switch info := fi.Sys().(type) {
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

	// on windows or generic fs.FS, see: https://www.mkssoftware.com/docs/man5/stat.5.asp
	return &File{
		Device: 0, Inode: 3,
		FileMode: fi.Mode(),
		UID:      0, GID: 0,
		NLink: 1, RDev: 0,
		ModifiedTime: fi.ModTime(),
	}
}
