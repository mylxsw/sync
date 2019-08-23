package server

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/mylxsw/coll"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
	"github.com/pkg/errors"
)

// SyncServer 同步服务端实现
type SyncServer struct {
	bufferSize int64
}

// NewSyncServer 创建一个文件同步服务
func NewSyncServer(bufferSize int64) *SyncServer {
	return &SyncServer{bufferSize: bufferSize}
}

// Download 文件下载服务
func (s *SyncServer) SyncFile(req *protocol.DownloadRequest, serv protocol.SyncService_SyncFileServer) error {
	f, err := os.Open(req.Filename)
	if err != nil {
		return errors.Wrapf(err, "can not open such file: %s", req.Filename)
	}

	defer f.Close()

	var writing = true
	buf := make([]byte, s.bufferSize)
	for writing {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				writing = false
				continue
			}

			return errors.Wrapf(err, "read file %s failed", req.Filename)
		}

		if err := serv.Send(&protocol.DownloadResponse{
			Content: buf[:n],
		}); err != nil {
			return errors.Wrap(err, "send file chunk failed")
		}
	}

	return nil
}

// Sync 文件元信息同步
func (s *SyncServer) SyncMeta(ctx context.Context, req *protocol.SyncRequest) (*protocol.SyncResponse, error) {
	matches, err := filepath.Glob(req.Path)
	if err != nil {
		return nil, errors.Wrap(err, "invalid glob expression")
	}

	files := make([]utils.File, 0)
	for _, f := range matches {
		ffs, err := utils.AllFiles(f)
		if err != nil {
			return nil, errors.Wrap(err, "read all files failed")
		}

		files = append(files, ffs...)
	}

	resp := protocol.SyncResponse{}
	if err := coll.Map(files, &resp.Files, func(f utils.File) *protocol.File {
		return &protocol.File{
			Path:     f.Path,
			Checksum: f.Checksum,
			Size:     f.Size,
			Type:     protocol.Type(protocol.Type_value[string(f.Type)]),
			Mode:     f.Mode,
			Uid:      f.UID,
			Gid:      f.GID,
			User:     f.User,
			Group:    f.Group,
		}
	}); err != nil {
		return nil, errors.Wrap(err, "convert []utils.File to resp.Files failed")
	}

	return &resp, nil
}
