package queue

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/rpc"
	"github.com/pkg/errors"
)

// FileSyncJob 文件同步任务
type FileSyncJob struct {
	group client.FileSyncGroup
}

// NewFileSyncJob 创建一个文件同步job
func NewFileSyncJob(group client.FileSyncGroup) *FileSyncJob {
	return &FileSyncJob{group: group}
}

func (job *FileSyncJob) Encode() []byte {
	res, _ := json.Marshal(job.group)
	return res
}

func (job *FileSyncJob) Decode(res []byte) {
	_ = json.Unmarshal(res, &job.group)
}

func (job *FileSyncJob) Handle(ctx context.Context, rpcFactory rpc.Factory) error {
	syncClient, err := rpcFactory.SyncClient(job.group.From, job.group.Token)
	if err != nil {
		return errors.Wrap(err, "create sync rpc client failed")
	}

	units := make([]client.SyncUnit, 0)
	for _, f := range job.group.Files {
		files, err := syncClient.SyncMeta(f)
		if err != nil {
			return errors.Wrap(err, "sync meta failed")
		}

		units = append(units, client.SyncUnit{
			Files:      files,
			FileToSync: f,
		})
	}

	// 分组前置任务
	for _, before := range job.group.Before {
		if before.Matched(units) {
			if err := before.Execute(units); err != nil {
				return errors.Wrap(err, "execute group before stage failed")
			}
		}
	}

	for _, g := range units {
		// 文件同步前置任务
		for _, before := range g.FileToSync.Before {
			if before.Matched([]client.SyncUnit{g}) {
				if err := before.Execute([]client.SyncUnit{g}); err != nil {
					return errors.Wrap(err, "execute before stage failed")
				}
			}
		}

		// 文件同步过程
		if err := syncClient.SyncFiles(g.Files, job.createSavePathGenerator(g.FileToSync), true); err != nil {
			return errors.Wrap(err, "file sync failed")
		}

		// 文件同步后置任务
		for _, after := range g.FileToSync.After {
			if after.Matched([]client.SyncUnit{g}) {
				if err := after.Execute([]client.SyncUnit{g}); err != nil {
					return errors.Wrap(err, "execute after stage failed")
				}
			}
		}
	}

	// 分组后置任务
	for _, after := range job.group.After {
		if after.Matched(units) {
			if err := after.Execute(units); err != nil {
				return errors.Wrap(err, "execute group before stage failed")
			}
		}
	}

	return nil
}

// createSavePathGenerator 创建一个文件保存路径生成器
func (job *FileSyncJob) createSavePathGenerator(fileToSync client.File) func(f *protocol.File) string {
	return func(f *protocol.File) string {
		return filepath.Join(fileToSync.Dest, strings.TrimPrefix(f.Path, fileToSync.Src))
	}
}
