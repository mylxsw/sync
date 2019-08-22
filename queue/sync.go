package queue

import (
	"context"
	"encoding/json"

	"github.com/mylxsw/sync/rpc"
	"github.com/pkg/errors"
)

// FileSyncJob 文件同步任务
type FileSyncJob struct {
	Path       string `json:"path"`
	ServerAddr string `json:"server_addr"`
	Token      string `json:"token"`
}

func (job *FileSyncJob) Encode() []byte {
	res, _ := json.Marshal(job)
	return res
}

func (job *FileSyncJob) Decode(res []byte) {
	_ = json.Unmarshal(res, job)
}

func (job *FileSyncJob) Handle(ctx context.Context, rpcFactory rpc.Factory) error {
	syncClient, err := rpcFactory.SyncClient(job.ServerAddr, job.Token)
	if err != nil {
		return errors.Wrap(err, "create sync rpc client failed")
	}

	if err := syncClient.Sync(job.Path); err != nil {
		return errors.Wrap(err, "sync failed")
	}

	return nil
}
