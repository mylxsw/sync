package client

import (
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/codingsince1985/checksum"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/coll"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
)

type FileSync interface {
	Sync(path string) error
}

// fileSyncClient 文件同步客户端
type fileSyncClient struct {
	client protocol.SyncServiceClient
}

// NewFileSync 创建一个文件同步客户端
func NewFileSync(client protocol.SyncServiceClient) FileSync {
	return &fileSyncClient{client: client}
}

// Sync 执行文件同步
func (fs *fileSyncClient) Sync(path string) error {
	resp, err := fs.client.Sync(context.TODO(), &protocol.SyncRequest{Path: path})
	if err != nil {
		return err
	}

	fs.applyFiles(resp.Files, protocol.Type_Directory, func(f *protocol.File, savedFilePath string) error {
		return os.MkdirAll(savedFilePath, os.FileMode(f.Mode))
	})
	fs.applyFiles(resp.Files, protocol.Type_Normal, fs.syncNormalFiles)
	fs.applyFiles(resp.Files, protocol.Type_Symlink, func(f *protocol.File, savedFilePath string) error {
		return os.Symlink(f.Symlink, savedFilePath)
	})

	return nil
}

// syncNormalFiles 同步普通文件
func (fs *fileSyncClient) syncNormalFiles(f *protocol.File, savedFilePath string) error {
	skipDownload := false
	if utils.FileExist(savedFilePath) {
		finger, _ := checksum.MD5sum(savedFilePath)
		if finger == f.Checksum {
			log.Debugf("skip file %s because it already exist", f.Path)
			skipDownload = true
		}
	}

	if !skipDownload {
		downloadResp, err := fs.client.Download(context.TODO(), &protocol.DownloadRequest{Filename: f.Path,})
		if err != nil {
			return err
		}

		if err := fs.writeFile(downloadResp, f, savedFilePath); err != nil {
			return err
		}
	}

	// checksum match confirm
	finger, _ := checksum.MD5sum(savedFilePath)
	if finger != f.Checksum {
		log.Errorf("file %s checksum not match, expect %s, but got %s", f.Path, f.Checksum, finger)
	}

	// file mode
	finfo, _ := os.Stat(savedFilePath)
	if finfo.Mode() != os.FileMode(f.Mode) {
		log.Infof("file mode changed for %s, %s -> %s", f.Path, finfo.Mode(), os.FileMode(f.Mode))
		if err := os.Chmod(savedFilePath, os.FileMode(f.Mode)); err != nil {
			log.Errorf("change file mode for %s failed: %s", f.Path, err)
		}
	}

	return nil
}

// applyFiles 批量处理指定类型的文件
func (fs *fileSyncClient) applyFiles(files []*protocol.File, fileType protocol.Type, cb func(f *protocol.File, savedFilePath string) error) {
	coll.MustNew(files).Filter(func(f *protocol.File) bool {
		return f.Type == fileType
	}).Each(func(f *protocol.File) {
		// log.Infof("PATH=%s，SIZE=%d，TYPE=%s，CHECKSUM=%s\n", f.Path, f.Size, f.Type.String(), f.Checksum)
		savedFilePath := filepath.Join("/tmp/", f.Path)

		if err := cb(f, savedFilePath); err != nil {
			log.Errorf("apply file %s failed: %s", f.Path, err)
			return
		}

		fs.syncFileOwner(savedFilePath, f)
	})
}

// syncFileOwner 同步文件属主
func (fs *fileSyncClient) syncFileOwner(dest string, f *protocol.File) {
	finfo, _ := os.Stat(dest)
	stat, _ := finfo.Sys().(*syscall.Stat_t)
	uid := -1
	gid := -1
	if f.User != "" {
		if u, err := user.Lookup(f.User); err == nil {
			if u.Uid != strconv.Itoa(int(stat.Uid)) {
				uid, _ = strconv.Atoi(u.Uid)
			}
		} else {
			if f.Uid != stat.Uid {
				uid = int(f.Uid)
			}
		}
	}

	if f.Group != "" {
		if g, err := user.LookupGroup(f.Group); err == nil {
			if g.Gid != strconv.Itoa(int(stat.Gid)) {
				gid, _ = strconv.Atoi(g.Gid)
			}
		} else {
			if f.Gid != stat.Gid {
				gid = int(f.Gid)
			}
		}
	}

	if uid != -1 || gid != -1 {
		if err := os.Chown(dest, uid, gid); err != nil {
			log.Errorf("chown for %s with uid=%d, gid=%d failed: %s", dest, uid, gid, err)
		}
	}
}

// writeFile 创建新文件
func (fs *fileSyncClient) writeFile(downloadResp protocol.SyncService_DownloadClient, f *protocol.File, savedFilePath string) error {
	log.Debugf("write file %s with mode=%s, size=%d ...", savedFilePath, os.FileMode(f.Mode), f.Size)

	saveFile, err := os.OpenFile(savedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(f.Mode))
	if err != nil {
		return err
	}

	defer saveFile.Close()

	total := 0
	for {
		recv, err := downloadResp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		cur, err := saveFile.Write(recv.Content)
		if err != nil {
			return err
		}

		total += cur
	}

	log.Infof("write file %s, size=%d OK", savedFilePath, total)
	return nil
}
