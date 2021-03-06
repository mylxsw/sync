package rpc

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/dustin/go-humanize"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Factory RPC 客户端创建工厂
type Factory interface {
	// SyncClient 创建一个文件同步客户端，使用完成之后不要忘记执行 Close 关闭连接
	SyncClient(endpoint string, token string) (FileSyncClient, error)
}

type factory struct{}

// NewFactory 创建一个客户端创建工厂
func NewFactory() Factory {
	return &factory{}
}

// SyncClient 创建一个文件同步客户端
func (factory *factory) SyncClient(endpoint string, token string) (FileSyncClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(), grpc.WithPerRPCCredentials(NewAuthAPI(token)))
	if err != nil {
		return nil, errors.Wrap(err, "can't dial to remote rpc server")
	}

	return NewFileSyncClient(protocol.NewSyncServiceClient(conn), func() {
		_ = conn.Close()
		log.Debugf("grpc connection closed")
	}), nil
}

// FileSyncClient 文件同步客户端接口
type FileSyncClient interface {
	// SyncMeta 同步文件元数据
	SyncMeta(fileToSync meta.File) ([]*protocol.File, error)
	// SyncDiff 同步文件比较，用于获取本次要同步的文件
	SyncDiff(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool) (meta.FileNeedSyncs, error)
	// SyncFiles 同步文件
	SyncFiles(fileNeedSyncs meta.FileNeedSyncs, stage *collector.Stage) error
	// RelayInfo 检查上游是否有变更
	RelayInfo(names []string) ([]*protocol.RelayFile, error)
	// Close close GRPC connection
	Close()
}

// fileSyncClient 文件同步客户端
type fileSyncClient struct {
	client  protocol.SyncServiceClient
	closeFn func()
}

// NewFileSyncClient 创建一个文件同步客户端
func NewFileSyncClient(client protocol.SyncServiceClient, closeFn func()) FileSyncClient {
	return &fileSyncClient{client: client, closeFn: closeFn,}
}

func (fs *fileSyncClient) SyncMeta(fileToSync meta.File) ([]*protocol.File, error) {
	resp, err := fs.client.SyncMeta(context.TODO(), &protocol.SyncRequest{Path: fileToSync.Src, Ignores: fileToSync.Ignores,}, grpc.MaxCallRecvMsgSize(math.MaxInt32))
	if err != nil {
		return nil, err
	}

	return resp.Files, nil
}

func (fs *fileSyncClient) SyncDiff(files []*protocol.File, savePath func(f *protocol.File) string, syncOwner bool) (meta.FileNeedSyncs, error) {
	var fileNeedSyncs = meta.FileNeedSyncs{Files: make([]meta.FileNeedSync, 0)}
	// 目录同步
	if err := fs.applyFiles(files, protocol.Type_Directory, fs.syncDirectoryInit(syncOwner, &fileNeedSyncs), savePath); err != nil {
		return fileNeedSyncs, err
	}

	// 文件同步
	if err := fs.applyFiles(files, protocol.Type_Normal, fs.syncNormalFileInit(syncOwner, &fileNeedSyncs), savePath); err != nil {
		return fileNeedSyncs, err
	}

	// 符号链接同步
	if err := fs.applyFiles(files, protocol.Type_Symlink, fs.syncSymlinkInit(syncOwner, &fileNeedSyncs), savePath); err != nil {
		return fileNeedSyncs, err
	}

	return fileNeedSyncs, nil
}

func (fs *fileSyncClient) SyncFiles(fileNeedSyncs meta.FileNeedSyncs, stage *collector.Stage) error {
	files := fileNeedSyncs.All()
	progress := stage.Progress(len(files))
	for _, ff := range files {
		// for security reason, not allow `/` and empty save path
		if ff.SaveFilePath == "" || ff.SaveFilePath == "/" {
			return fmt.Errorf("security: sync path %s not allowed", ff.SaveFilePath)
		}

		if ff.SyncFile {
			switch ff.Type {
			case protocol.Type_Normal:
				if err := fs.syncNormalFiles(ff.RemoteFile, ff.SaveFilePath, stage); err != nil {
					return fmt.Errorf("sync file %s -> %s but failed: %s", ff.RemoteFile.Path, ff.SaveFilePath, err)
				}
			case protocol.Type_Directory:
				if err := fs.createDirectory(ff.SaveFilePath, ff.RemoteFile, stage); err != nil {
					return err
				}
			case protocol.Type_Symlink:
				if err := os.Symlink(ff.RemoteFile.Symlink, ff.SaveFilePath); err != nil {
					return fmt.Errorf("create symlink %s -> %s, but failed: %s", ff.SaveFilePath, ff.RemoteFile.Symlink, err)
				}

				stage.Info(fmt.Sprintf("ln %s -> %s", ff.SaveFilePath, ff.RemoteFile.Symlink))
			}
		}

		if ff.Chmod {
			if err := os.Chmod(ff.SaveFilePath, os.FileMode(ff.RemoteFile.Mode)); err != nil {
				stage.Error(fmt.Sprintf(
					"chmod %s (%s) failed: %s",
					ff.SaveFilePath,
					os.FileMode(ff.RemoteFile.Mode),
					err,
				))
			} else {
				stage.Info(fmt.Sprintf("chmod %s (%s)", ff.SaveFilePath, os.FileMode(ff.RemoteFile.Mode)))
			}
		}

		if ff.SyncOwner {
			if err := fs.syncFileOwner(ff.SaveFilePath, ff.RemoteFile); err != nil {
				stage.Error(err.Error())
			}
		}

		if ff.Delete {
			if err := os.RemoveAll(ff.SaveFilePath); err != nil {
				_ = stage.Errorf("rm -fr %s failed: %s", ff.SaveFilePath, err)
			} else {
				stage.Infof("rm -fr %s", ff.SaveFilePath)
			}
		}

		progress.Add(1)
	}
	return nil
}

func (fs *fileSyncClient) syncSymlinkInit(syncOwner bool, fileNeedSyncs *meta.FileNeedSyncs) func(f *protocol.File, savedFilePath string) error {
	return func(f *protocol.File, savedFilePath string) error {
		fileNeedSync := meta.FileNeedSync{
			SaveFilePath: savedFilePath,
			Type:         protocol.Type_Symlink,
			RemoteFile:   f,
		}

		if fs.needCreateSymlink(savedFilePath, f) {
			fileNeedSync.SyncFile = true
		} else if fs.needChmod(f, savedFilePath) {
			fileNeedSync.Chmod = true
		}

		if syncOwner && fs.needSyncFileOwner(savedFilePath, f) {
			fileNeedSync.SyncOwner = true
		}

		if fileNeedSync.NeedSync() {
			fileNeedSyncs.Add(fileNeedSync)
		}

		return nil
	}
}

func (fs *fileSyncClient) syncNormalFileInit(syncOwner bool, fileNeedSyncs *meta.FileNeedSyncs) func(f *protocol.File, savedFilePath string) error {
	return func(f *protocol.File, savedFilePath string) error {
		fileNeedSync := meta.FileNeedSync{
			SaveFilePath: savedFilePath,
			Type:         protocol.Type_Normal,
			RemoteFile:   f,
		}

		if fs.needSyncNormalFile(f, savedFilePath) {
			fileNeedSync.SyncFile = true
		} else if fs.needChmod(f, savedFilePath) {
			fileNeedSync.Chmod = true
		}

		if syncOwner && fs.needSyncFileOwner(savedFilePath, f) {
			fileNeedSync.SyncOwner = true
		}

		if fileNeedSync.NeedSync() {
			fileNeedSyncs.Add(fileNeedSync)
		}

		return nil
	}
}

func (fs *fileSyncClient) syncDirectoryInit(syncOwner bool, fileNeedSyncs *meta.FileNeedSyncs) func(f *protocol.File, savedFilePath string) error {
	return func(f *protocol.File, savedFilePath string) error {
		fileNeedSync := meta.FileNeedSync{
			SaveFilePath: savedFilePath,
			Type:         protocol.Type_Directory,
			RemoteFile:   f,
		}

		if fs.needCreateDirectory(savedFilePath, f) {
			fileNeedSync.SyncFile = true
		} else if fs.needChmod(f, savedFilePath) {
			fileNeedSync.Chmod = true
		}

		if syncOwner && fs.needSyncFileOwner(savedFilePath, f) {
			fileNeedSync.SyncOwner = true
		}

		if fileNeedSync.NeedSync() {
			fileNeedSyncs.Add(fileNeedSync)
		}

		return nil
	}
}

func (fs *fileSyncClient) createDirectory(savedFilePath string, f *protocol.File, stage *collector.Stage) error {
	if err := os.MkdirAll(savedFilePath, os.FileMode(f.Mode)); err != nil {
		return fmt.Errorf("create directory %s with permission %s, but failed: %s", savedFilePath, os.FileMode(f.Mode), err)
	}
	stage.Info(fmt.Sprintf("mkdir %s (%s)", savedFilePath, os.FileMode(f.Mode)))
	return nil
}

// needCreateDirectory 返回是否需要创建目录，如果目录存在，但是权限不一样，会自动修正权限
func (fs *fileSyncClient) needCreateDirectory(savedFilePath string, f *protocol.File) bool {
	if utils.FileExist(savedFilePath) {
		info, err := os.Lstat(savedFilePath)
		if err != nil {
			log.Errorf("get file %s info failed: %s", savedFilePath, err)
		} else {
			if !info.IsDir() {
				log.Warningf("file %s is not a directory, we will remove it and recreate as a directory", savedFilePath)
				if err := os.Remove(savedFilePath); err != nil {
					log.Errorf("delete file %s failed: %s", savedFilePath, err)
				}

				return true
			}

			if info.Mode() != os.FileMode(f.Mode) {
				// 修改目录权限
				if err := os.Chmod(savedFilePath, os.FileMode(f.Mode)); err != nil {
					log.Errorf("can not change file mode for %s: %s", savedFilePath, err)
				}
			}
		}

		return false
	}

	return true
}

// needCreateSymlink 返回是否需要创建符号链接，如果符号链接与要同步的数据不一致，则自动删除
func (fs *fileSyncClient) needCreateSymlink(savedFilePath string, f *protocol.File) bool {
	skipFile := false
	if utils.FileExist(savedFilePath) {
		info, err := os.Lstat(savedFilePath)
		if err != nil {
			log.Errorf("get file %s info failed: %s", savedFilePath, err)
		} else {
			if info.Mode()&os.ModeSymlink != 0 {
				link, _ := os.Readlink(savedFilePath)
				if link == f.Symlink {
					skipFile = true
				} else {
					if err := os.Remove(savedFilePath); err != nil {
						log.Errorf("failed to remove symlink %s: %s", savedFilePath, err)
					}
				}
			}
		}
	}
	return !skipFile
}

// needSyncNormalFile 判断是否需要同步普通文件
func (fs *fileSyncClient) needSyncNormalFile(f *protocol.File, savedFilePath string) bool {
	if utils.FileExist(savedFilePath) {
		finger, _ := checksum.MD5sum(savedFilePath)
		if finger == f.Checksum {
			return false
		}
	}

	return true
}

// needChmod 判断是否执行 chmod
func (fs *fileSyncClient) needChmod(f *protocol.File, savedFilePath string) bool {
	finfo, _ := os.Stat(savedFilePath)
	return finfo.Mode() != os.FileMode(f.Mode)
}

// syncNormalFiles 同步普通文件
func (fs *fileSyncClient) syncNormalFiles(f *protocol.File, savedFilePath string, stage *collector.Stage) error {
	startTs := time.Now()
	downloadResp, err := fs.client.SyncFile(context.TODO(), &protocol.DownloadRequest{Filename: filepath.Join(f.Base, f.Path),})
	if err != nil {
		return err
	}

	if err := fs.writeFile(downloadResp, f, savedFilePath); err != nil {
		return err
	}

	remoteFileName := f.Path
	if remoteFileName == "" {
		remoteFileName = filepath.Base(savedFilePath)
	}

	stage.Info(fmt.Sprintf(
		"copy %s -> %s (elapse=%v size=%s)",
		remoteFileName,
		savedFilePath,
		time.Now().Sub(startTs),
		humanize.Bytes(uint64(f.Size)),
	))

	// checksum match confirm
	finger, _ := checksum.MD5sum(savedFilePath)
	if finger != f.Checksum {
		stage.Error(fmt.Sprintf(
			"copy %s -> %s (checksum not match: %s != %s)",
			remoteFileName,
			savedFilePath,
			finger,
			f.Checksum,
		))
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

// needSyncFileOwner 是否需要同步文件属主
func (fs *fileSyncClient) needSyncFileOwner(dest string, f *protocol.File) bool {
	finfo, err := os.Stat(dest)
	if err != nil {
		return true
	}

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

	return uid != -1 && gid != -1
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

	log.Debugf("write file %s, mode=%s, size=%s OK", savedFilePath, os.FileMode(f.Mode), humanize.Bytes(uint64(total)))
	return nil
}

func (fs *fileSyncClient) RelayInfo(names []string) ([]*protocol.RelayFile, error) {
	resp, err := fs.client.RelayInfo(context.TODO(), &protocol.RelayRequest{Names: names,})
	if err != nil {
		return nil, err
	}

	return resp.Files, err
}

func (fs *fileSyncClient) Close() {
	if fs.closeFn != nil {
		fs.closeFn()
	}
}
