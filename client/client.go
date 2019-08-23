package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
)

// FileSyncClient 文件同步客户端接口
type FileSyncClient interface {
	// SyncMeta 同步文件元数据
	SyncMeta(fileToSync File) ([]*protocol.File, error)
	// SyncFiles 同步文件
	SyncFiles(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool, stage *collector.Stage) error
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
func (fs *fileSyncClient) SyncFiles(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool, stage *collector.Stage) error {
	// 目录同步
	if err := fs.applyFiles(files, protocol.Type_Directory, func(f *protocol.File, savedFilePath string) error {
		if err := os.MkdirAll(savedFilePath, os.FileMode(f.Mode)); err != nil {
			return fmt.Errorf("create directory %s with permission %s, but failed: %s", savedFilePath, os.FileMode(f.Mode), err)
		}

		stage.Log(fmt.Sprintf("create directory %s with permission %s", savedFilePath, os.FileMode(f.Mode)))

		if syncOwner {
			if err := fs.syncFileOwner(savedFilePath, f); err != nil {
				stage.Error(err.Error())
			}
		}

		return nil
	}, savePath); err != nil {
		return err
	}

	// 文件同步
	if err := fs.applyFiles(files, protocol.Type_Normal, func(f *protocol.File, savedFilePath string) error {
		if err := fs.syncNormalFiles(f, savedFilePath, stage); err != nil {
			return fmt.Errorf("sync file %s -> %s but failed: %s", f.Path, savedFilePath, err)
		}

		if syncOwner {
			if err := fs.syncFileOwner(savedFilePath, f); err != nil {
				stage.Error(err.Error())
			}
		}

		return nil
	}, savePath); err != nil {
		return err
	}

	// 符号链接同步
	if err := fs.applyFiles(files, protocol.Type_Symlink, func(f *protocol.File, savedFilePath string) error {
		if err := os.Symlink(f.Symlink, savedFilePath); err != nil {
			return fmt.Errorf("create symlink %s -> %s, but failed: %s", f.Path, savedFilePath, err)
		}

		stage.Log(fmt.Sprintf("create symlink %s -> %s", f.Path, savedFilePath))

		if syncOwner {
			if err := fs.syncFileOwner(savedFilePath, f); err != nil {
				stage.Error(err.Error())
			}
		}

		return nil
	}, savePath); err != nil {
		return err
	}

	return nil
}

// syncNormalFiles 同步普通文件
func (fs *fileSyncClient) syncNormalFiles(f *protocol.File, savedFilePath string, stage *collector.Stage) error {
	skipDownload := false
	if utils.FileExist(savedFilePath) {
		finger, _ := checksum.MD5sum(savedFilePath)
		if finger == f.Checksum {
			log.Debugf("skip file %s because it already exist", f.Path)
			skipDownload = true
		}
	}

	if !skipDownload {
		startTs := time.Now()
		downloadResp, err := fs.client.SyncFile(context.TODO(), &protocol.DownloadRequest{Filename: f.Path,})
		if err != nil {
			return err
		}

		if err := fs.writeFile(downloadResp, f, savedFilePath); err != nil {
			return err
		}

		stage.Log(fmt.Sprintf("sync file %s -> %s finished, elapse %v， size=%d", f.Path, savedFilePath, time.Now().Sub(startTs), f.Size))
	}

	// checksum match confirm
	finger, _ := checksum.MD5sum(savedFilePath)
	if finger != f.Checksum {
		log.Errorf("file %s checksum not match, expect %s, but got %s", f.Path, f.Checksum, finger)
		stage.Error(fmt.Sprintf(
			"sync file %s -> %s finished, but checksum not match: %s != %s",
			f.Path,
			savedFilePath,
			finger,
			f.Checksum,
		))
	}

	// file mode
	finfo, _ := os.Stat(savedFilePath)
	if finfo.Mode() != os.FileMode(f.Mode) {
		log.Infof("file mode changed for %s, %s -> %s", f.Path, finfo.Mode(), os.FileMode(f.Mode))
		if err := os.Chmod(savedFilePath, os.FileMode(f.Mode)); err != nil {
			log.Errorf("change file mode for %s failed: %s", f.Path, err)
			stage.Error(fmt.Sprintf(
				"change file mode for %s from %s to %s, but failed: %s",
				savedFilePath,
				finfo.Mode(),
				os.FileMode(f.Mode),
				err,
			))
		} else {
			stage.Log(fmt.Sprintf("change file mode for %s from %s to %s", savedFilePath, finfo.Mode(), os.FileMode(f.Mode)))
		}
	}

	return nil
}

// applyFiles 批量处理指定类型的文件
func (fs *fileSyncClient) applyFiles(
	files []*protocol.File,
	fileType protocol.Type,
	cb func(f *protocol.File, savedFilePath string) error,
	filePath func(f *protocol.File) string,
) error {
	for _, item := range files {
		if item.Type != fileType {
			continue
		}

		savedFilePath := filePath(item)

		if err := cb(item, savedFilePath); err != nil {
			log.Errorf("apply file %s failed: %s", item.Path, err)
			return err
		}
	}

	return nil
}

// syncFileOwner 同步文件属主
func (fs *fileSyncClient) syncFileOwner(dest string, f *protocol.File) error {
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
			errMsg := fmt.Sprintf("chown for %s with uid=%d, gid=%d failed: %s", dest, uid, gid, err)
			log.Error(errMsg)
			return errors.New(errMsg)
		}
	}

	return nil
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
