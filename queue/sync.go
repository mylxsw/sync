package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/utils"
	"github.com/pkg/errors"
)

// FileSyncJob 文件同步任务
type FileSyncJob struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Payload   client.FileSyncGroup `json:"payload"`
	CreatedAt time.Time            `json:"created_at"`
}

// NewFileSyncJob 创建一个文件同步job
func NewFileSyncJob(group client.FileSyncGroup) *FileSyncJob {
	return &FileSyncJob{
		ID:        utils.UUID(),
		Name:      "file-sync",
		Payload:   group,
		CreatedAt: time.Now(),
	}
}

func (job *FileSyncJob) Encode() []byte {
	res, _ := json.Marshal(job)
	return res
}

func (job *FileSyncJob) Decode(res []byte) {
	_ = json.Unmarshal(res, &job)
}

func (job *FileSyncJob) Handle(ctx context.Context, rpcFactory rpc.Factory, col *collector.Collector) error {
	syncClient, err := rpcFactory.SyncClient(job.Payload.From, job.Payload.Token)
	if err != nil {
		return errors.Wrap(err, "create sync rpc client failed")
	}

	stageSyncMeta := col.Stage("sync-meta")

	units := make([]client.SyncUnit, 0)
	for i, f := range job.Payload.Files {
		files, err := syncClient.SyncMeta(f)
		if err != nil {
			stageSyncMeta.Error(fmt.Sprintf("#%d sync meta failed: %s", i, err))
			return errors.Wrap(err, "sync meta failed")
		}

		units = append(units, client.SyncUnit{
			Files:      files,
			FileToSync: f,
		})

		stageSyncMeta.Log(fmt.Sprintf("#%d has %d files", i, len(files)))
	}

	// 分组前置任务
	stageGroupBefore := col.Stage("group-before")
	for i, before := range job.Payload.Before {
		if before.Matched(units) {
			if err := before.Execute(units); err != nil {
				stageGroupBefore.Error(fmt.Sprintf("#%d matched, but execute failed: %s", i, err))
				return errors.Wrap(err, "execute Payload before stage failed")
			}

			stageGroupBefore.Log(fmt.Sprintf("#%d matched and ok", i))
		}
	}

	for i, g := range units {
		// 文件同步前置任务
		stageSyncBefore := col.Stage(fmt.Sprintf("sync-before-#%d", i))
		for j, before := range g.FileToSync.Before {
			if before.Matched([]client.SyncUnit{g}) {
				if err := before.Execute([]client.SyncUnit{g}); err != nil {
					stageSyncBefore.Error(fmt.Sprintf("#%d matched, but execute failed: %s", j, err))
					return errors.Wrap(err, "execute before stage failed")
				}

				stageSyncBefore.Log(fmt.Sprintf("#%d matched and ok", j))
			}
		}

		// 文件同步过程
		stageSync := col.Stage(fmt.Sprintf("sync-files-#%d", i))
		if err := syncClient.SyncFiles(g.Files, job.createSavePathGenerator(g.FileToSync), true, stageSync); err != nil {
			stageSync.Error(err.Error())
			return errors.Wrap(err, "file sync failed")
		}

		// 文件同步后置任务
		stageSyncAfter := col.Stage(fmt.Sprintf("sync-after-#%d", i))
		for j, after := range g.FileToSync.After {
			if after.Matched([]client.SyncUnit{g}) {
				if err := after.Execute([]client.SyncUnit{g}); err != nil {
					stageSyncAfter.Error(fmt.Sprintf("#%d matched, but execute failed: %s", j, err))
					return errors.Wrap(err, "execute after stage failed")
				}

				stageSyncAfter.Log(fmt.Sprintf("#%d matched and ok", j))
			}
		}
	}

	// 分组后置任务
	stageGroupAfter := col.Stage("group-after")
	for i, after := range job.Payload.After {
		if after.Matched(units) {
			if err := after.Execute(units); err != nil {
				stageGroupAfter.Error(fmt.Sprintf("#%d matched, but execute failed: %s", i, err))
				return errors.Wrap(err, "execute Payload before stage failed")
			}

			stageGroupAfter.Log(fmt.Sprintf("#%d matched and ok", i))
		}
	}

	return nil
}

// createSavePathGenerator 创建一个文件保存路径生成器
func (job *FileSyncJob) createSavePathGenerator(fileToSync client.File, ) func(f *protocol.File) string {
	return func(f *protocol.File) string {
		return filepath.Join(fileToSync.Dest, strings.TrimPrefix(f.Path, fileToSync.Src))
	}
}
