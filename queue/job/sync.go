package job

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/queue/action"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/storage"
	"github.com/mylxsw/sync/utils"
	"github.com/pkg/errors"
)

// FileSyncJob is a job for file sync
type FileSyncJob struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Payload   meta.FileSyncGroup `json:"payload"`
	CreatedAt time.Time          `json:"created_at"`

	cc            *container.Container
	syncSetting   *meta.GlobalFileSyncSetting
	actionFactory action.Factory
}

// NewFileSyncJob create a FileSyncJob
func NewFileSyncJob(group meta.FileSyncGroup) *FileSyncJob {
	return &FileSyncJob{
		ID:        utils.UUID(),
		Name:      "file-sync",
		Payload:   group,
		CreatedAt: time.Now(),
	}
}

func (job *FileSyncJob) Init(settingFactory storage.SettingFactory, actionFactory action.Factory) {
	job.actionFactory = actionFactory
	syncSettingData, err := settingFactory.Namespace(storage.GlobalNamespace).Get(storage.SyncActionSetting)
	if err != nil {
		if err != storage.ErrNoSuchSetting {
			log.Errorf("load global sync setting failed: %s", err)
		}

		job.syncSetting = meta.NewGlobalFileSyncSetting()
		return
	}

	setting := meta.GlobalFileSyncSetting{}
	if err := setting.Decode(syncSettingData); err != nil {
		log.Errorf("decode global sync setting failed: %s", err)
		job.syncSetting = meta.NewGlobalFileSyncSetting()
		return
	}

	job.syncSetting = &setting
}

func (job *FileSyncJob) BeforeActions() []meta.SyncAction {
	return append(job.Payload.Before, job.syncSetting.Before...)
}

func (job *FileSyncJob) AfterActions() []meta.SyncAction {
	return append(job.Payload.After, job.syncSetting.After...)
}

func (job *FileSyncJob) ErrorActions() []meta.SyncAction {
	return append(job.Payload.Error, job.syncSetting.Errors...)
}

func (job *FileSyncJob) RemoteServer() string {
	if job.Payload.From != "" {
		return job.Payload.From
	}

	return job.syncSetting.From
}

func (job *FileSyncJob) RemoteServerToken() string {
	if job.Payload.Token != "" {
		return job.Payload.Token
	}

	return job.syncSetting.Token
}

func (job *FileSyncJob) Encode() []byte {
	res, _ := json.Marshal(job)
	return res
}

func (job *FileSyncJob) Decode(res []byte) {
	_ = json.Unmarshal(res, &job)
}

func (job *FileSyncJob) Handle(ctx context.Context, rpcFactory rpc.Factory, col *collector.Collector) error {
	if err := job.handle(ctx, rpcFactory, col); err != nil {
		errActions := job.ErrorActions()
		if len(errActions) > 0 {
			stage := col.Stage("errors")
			matchData := action.NewSyncMatchData(job.ID, job.Payload, []meta.SyncUnit{}, meta.FileNeedSyncs{}, err)
			for _, ac := range errActions {
				if act := job.actionFactory.Action(&ac, matchData); act != nil {
					if err := act.Execute(stage); err != nil {
						log.WithFields(log.Fields{
							"act": act,
						}).Errorf("execute error stage failed: %s", err)
					}
				}
			}
		}

		return err
	}

	return nil
}

func (job *FileSyncJob) handle(ctx context.Context, rpcFactory rpc.Factory, col *collector.Collector) error {
	syncClient, err := rpcFactory.SyncClient(job.RemoteServer(), job.RemoteServerToken())
	if err != nil {
		return errors.Wrap(err, "create sync rpc client failed")
	}

	// sync file metas
	units, err := job.syncMeta(col, syncClient)
	if err != nil {
		return err
	}

	// sync before
	if err := job.groupBefore(col, units); err != nil {
		return err
	}

	// syncing
	if err := job.fileSync(units, col, syncClient); err != nil {
		return err
	}

	// sync after
	if err := job.groupAfter(col, units); err != nil {
		return err
	}

	return nil
}

// fileSync execute file sync progress
func (job *FileSyncJob) fileSync(units []meta.SyncUnit, col *collector.Collector, syncClient client.FileSyncClient) error {
	for i, g := range units {

		// file diff, only sync changed files
		stageSync := col.Stage(fmt.Sprintf("sync-files-#%d", i))
		fileNeedSyncs, err := syncClient.SyncDiff(g.Files, job.createSavePathGenerator(g.FileToSync), true)
		if err != nil {
			stageSync.Errorf("sync file diff failed: %s", err.Error())
			return errors.Wrap(err, "file sync diff failed")
		}

		// file sync before
		stageSyncBefore := col.Stage(fmt.Sprintf("sync-before-#%d", i))
		for j, before := range g.FileToSync.Before {
			matchData := action.NewSyncMatchData(job.ID, job.Payload, []meta.SyncUnit{g}, fileNeedSyncs, nil)
			if beforeAct := job.actionFactory.Action(&before, matchData); beforeAct != nil {
				if err := beforeAct.Execute(stageSyncBefore); err != nil {
					stageSyncBefore.Error(fmt.Sprintf("#%d matched, but execute failed: %s", j, err))
					return errors.Wrap(err, "execute before stage failed")
				}

				stageSyncBefore.Info(fmt.Sprintf("#%d matched and ok", j))
			}
		}

		// real file sync progress
		if err := syncClient.SyncFiles(fileNeedSyncs, stageSync); err != nil {
			stageSync.Errorf("sync file failed: %s", err.Error())

			if len(g.FileToSync.Error) > 0 {
				errorStage := col.Stage(fmt.Sprintf("sync-errors-#%d", i))
				matchData := action.NewSyncMatchData(job.ID, job.Payload, []meta.SyncUnit{}, fileNeedSyncs, err)
				for _, ac := range g.FileToSync.Error {
					if act := job.actionFactory.Action(&ac, matchData); act != nil {
						if err := act.Execute(errorStage); err != nil {
							log.WithFields(log.Fields{
								"act": act,
							}).Errorf("execute error stage for %d failed: %s", i, err)
						}
					}
				}
			}

			if g.FileToSync.SkipWhenError {
				continue
			}

			return errors.Wrap(err, "file sync failed")
		}

		// file sync after
		stageSyncAfter := col.Stage(fmt.Sprintf("sync-after-#%d", i))
		for j, after := range g.FileToSync.After {
			matchData := action.NewSyncMatchData(job.ID, job.Payload, []meta.SyncUnit{g}, fileNeedSyncs, nil)
			if act := job.actionFactory.Action(&after, matchData); act != nil {
				if err := act.Execute(stageSyncAfter); err != nil {
					stageSyncAfter.Error(fmt.Sprintf("#%d matched, but execute failed: %s", j, err))
					return errors.Wrap(err, "execute after stage failed")
				}

				stageSyncAfter.Info(fmt.Sprintf("#%d matched and ok", j))
			}
		}
	}

	return nil
}

func (job *FileSyncJob) groupAfter(col *collector.Collector, units []meta.SyncUnit) error {
	stageGroupAfter := col.Stage("group-after")
	for i, after := range job.AfterActions() {
		matchData := action.NewSyncMatchData(job.ID, job.Payload, units, meta.FileNeedSyncs{}, nil)
		if act := job.actionFactory.Action(&after, matchData); act != nil {
			if err := act.Execute(stageGroupAfter); err != nil {
				stageGroupAfter.Error(fmt.Sprintf("#%d matched, but execute failed: %s", i, err))
				return errors.Wrap(err, "execute Payload before stage failed")
			}

			stageGroupAfter.Info(fmt.Sprintf("#%d matched and ok", i))
		}
	}

	return nil
}

func (job *FileSyncJob) groupBefore(col *collector.Collector, units []meta.SyncUnit) error {
	stageGroupBefore := col.Stage("group-before")
	for i, before := range job.BeforeActions() {
		matchData := action.NewSyncMatchData(job.ID, job.Payload, units, meta.FileNeedSyncs{}, nil)
		if act := job.actionFactory.Action(&before, matchData); act != nil {
			if err := act.Execute(stageGroupBefore); err != nil {
				stageGroupBefore.Error(fmt.Sprintf("#%d matched, but execute failed: %s", i, err))
				return errors.Wrap(err, "execute Payload before stage failed")
			}

			stageGroupBefore.Info(fmt.Sprintf("#%d matched and ok", i))
		}
	}

	return nil
}

func (job *FileSyncJob) syncMeta(col *collector.Collector, syncClient client.FileSyncClient) ([]meta.SyncUnit, error) {
	stageSyncMeta := col.Stage("sync-meta")
	units := make([]meta.SyncUnit, 0)
	for i, f := range job.Payload.Files {
		files, err := syncClient.SyncMeta(f)
		if err != nil {
			stageSyncMeta.Error(fmt.Sprintf("#%d sync meta failed: %s", i, err))
			return nil, errors.Wrap(err, "sync meta failed")
		}

		units = append(units, meta.SyncUnit{
			Files:      files,
			FileToSync: f,
		})

		stageSyncMeta.Info(fmt.Sprintf("#%d has %d files", i, len(files)))
	}
	return units, nil
}

// createSavePathGenerator create a file save path generator
func (job *FileSyncJob) createSavePathGenerator(fileToSync meta.File) func(f *protocol.File) string {
	return func(f *protocol.File) string {
		return filepath.Join(fileToSync.Dest, f.Path)
	}
}
