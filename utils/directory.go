package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/codingsince1985/checksum"
	"github.com/mylxsw/coll"
)

type FileType string

const (
	FileTypeNormal    FileType = "Normal"
	FileTypeDirectory FileType = "Directory"
	FileTypeSymlink   FileType = "Symlink"
)

type File struct {
	Path     string
	Checksum string
	Size     int64
	Type     FileType
	Symlink  string
	Mode     uint32
	UID      uint32
	User     string
	GID      uint32
	Group    string
	Base     string
}

// AllFiles 返回目录/文件下所有的目录/文件
func AllFiles(dir string) ([]File, error) {
	files := make([]File, 0)
	workdir, _ := filepath.EvalSymlinks(dir)
	if err := filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		file := File{
			Path: path,
			Size: info.Size(),
			Mode: uint32(info.Mode()),
			Base: dir,
		}

		stat, ok := info.Sys().(*syscall.Stat_t)
		if ok {
			file.UID = stat.Uid
			if u, err := user.LookupId(strconv.Itoa(int(stat.Uid))); err == nil {
				file.User = u.Username
			}

			file.GID = stat.Gid
			if g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid))); err == nil {
				file.Group = g.Name
			}
		}

		fileMode := info.Mode()
		if fileMode&os.ModeSymlink != 0 {
			file.Type = FileTypeSymlink
			file.Symlink, _ = os.Readlink(path)
		} else if fileMode&os.ModeDir != 0 {
			file.Type = FileTypeDirectory
		} else if fileMode&os.ModeType == 0 {
			file.Type = FileTypeNormal
		} else {
			return nil
		}

		if file.Type == FileTypeNormal {
			file.Checksum, _ = checksum.MD5sum(path)
		}

		files = append(files, file)
		return nil
	}); err != nil {
		return files, err
	}

	_ = coll.Map(files, &files, func(f File) File {
		f.Path = strings.TrimLeft(strings.TrimPrefix(f.Path, workdir), "/")
		return f
	})

	return files, nil
}
