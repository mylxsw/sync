package client

import (
	"context"
	"io"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/codingsince1985/checksum"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/coll"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
)

// FileSyncClient 文件同步客户端接口
type FileSyncClient interface {
	// SyncMeta 同步文件元数据
	SyncMeta(fileToSync File) ([]*protocol.File, error)
	// SyncFiles 同步文件
	SyncFiles(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool) error
}

// fileSyncClient 文件同步客户端
type fileSyncClient struct {
	client protocol.SyncServiceClient
}

// NewFileSyncClient 创建一个文件同步客户端
func NewFileSyncClient(client protocol.SyncServiceClient) FileSyncClient {
	return &fileSyncClient{client: client}
}

func (fs *fileSyncClient) SyncMeta(fileToSync File) ([]*protocol.File, error) {
	resp, err := fs.client.SyncMeta(context.TODO(), &protocol.SyncRequest{Path: fileToSync.Src})
	if err != nil {
		return nil, err
	}

	return resp.Files, nil
}

// Sync 执行文件同步
func (fs *fileSyncClient) SyncFiles(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool) error {
	fs.applyFiles(files, protocol.Type_Directory, func(f *protocol.File, savedFilePath string) error {
		if err := os.MkdirAll(savedFilePath, os.FileMode(f.Mode)); err != nil {
			return err
		}

		if syncOwner {
			fs.syncFileOwner(savedFilePath, f)
		}

		return nil
	}, savePath)

	fs.applyFiles(files, protocol.Type_Normal, func(f *protocol.File, savedFilePath string) error {
		if err := fs.syncNormalFiles(f, savedFilePath); err != nil {
			return err
		}

		if syncOwner {
			fs.syncFileOwner(savedFilePath, f)
		}

		return nil
	}, savePath)

	fs.applyFiles(files, protocol.Type_Symlink, func(f *protocol.File, savedFilePath string) error {
		if err := os.Symlink(f.Symlink, savedFilePath); err != nil {
			return err
		}

		if syncOwner {
			fs.syncFileOwner(savedFilePath, f)
		}

		return nil
	}, savePath)

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
		downloadResp, err := fs.client.SyncFile(context.TODO(), &protocol.DownloadRequest{Filename: f.Path,})
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
func (fs *fileSyncClient) applyFiles(files []*protocol.File, fileType protocol.Type, cb func(f *protocol.File, savedFilePath string) error, filePath func(f *protocol.File) string) {
	coll.MustNew(files).Filter(func(f *protocol.File) bool {
		return f.Type == fileType
	}).Each(func(f *protocol.File) {
		// log.Infof("PATH=%s，SIZE=%d，TYPE=%s，CHECKSUM=%s\n", f.Src, f.Size, f.Type.String(), f.Checksum)
		savedFilePath := filePath(f)

		if err := cb(f, savedFilePath); err != nil {
			log.Errorf("apply file %s failed: %s", f.Path, err)
			return
		}
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
func (fs *fileSyncClient) writeFile(downloadResp protocol.SyncService_SyncFileClient, f *protocol.File, savedFilePath string) error {
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
