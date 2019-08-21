package server

import (
	"context"
	"io"
	"os"

	"github.com/mylxsw/coll"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/utils"
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
func (s *SyncServer) Download(req *protocol.DownloadRequest, serv protocol.SyncService_DownloadServer) error {
	f, err := os.Open(req.Filename)
	if err != nil {
		return err
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

			return err
		}

		if err := serv.Send(&protocol.DownloadResponse{
			Content: buf[:n],
		}); err != nil {
			return err
		}
	}

	return nil
}

// Sync 文件元信息同步
func (s *SyncServer) Sync(ctx context.Context, req *protocol.SyncRequest) (*protocol.SyncResponse, error) {
	files, err := utils.AllFiles(req.Path)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &resp, nil
}
