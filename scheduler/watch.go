package scheduler

import (
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/storage"
)

type Watcher struct {
	cc *container.Container

	clientFactory rpc.Factory
	settingStore  storage.SettingStore
	statusStore   storage.JobStatusStore
	syncQueue     queue.SyncQueue
}

func NewWatcher(cc *container.Container) *Watcher {
	watcher := Watcher{cc: cc}
	cc.MustResolve(func(cf rpc.Factory, factory storage.SettingFactory, statusStore storage.JobStatusStore, syncQueue queue.SyncQueue) {
		watcher.clientFactory = cf
		watcher.settingStore = factory.Namespace(storage.GlobalNamespace)
		watcher.statusStore = statusStore
		watcher.syncQueue = syncQueue
	})
	return &watcher
}

func (watcher *Watcher) Handle() {
	data, err := watcher.settingStore.Get(storage.SyncActionSetting)
	if err != nil {
		log.Errorf("get sync setting failed: %s", err)
		return
	}

	setting := meta.GlobalFileSyncSetting{}
	if err := setting.Decode(data); err != nil {
		log.Errorf("parse sync setting failed: %s", err)
		return
	}

	if len(setting.Watches) == 0 {
		return
	}

	for _, sw := range setting.Watches {
		from, token := sw.From, sw.Token
		if from == "" {
			from, token = setting.From, setting.Token
		}

		c, err := watcher.clientFactory.SyncClient(from, token)
		if err != nil {
			log.WithFields(log.Fields{
				"from":  sw.From,
				"token": sw.Token,
			}).Errorf("create sync client failed: %s", err)
			continue
		}

		watchFiles, err := c.Watch(sw.Names)
		if err != nil {
			log.WithFields(log.Fields{
				"names": sw.Names,
			}).Errorf("rpc watch files failed: %s", err)
			continue
		}

		for _, wf := range watchFiles {
			if wf.LastStatus != "ok" {
				continue
			}

			lastUpdate, _ := time.Parse(time.RFC3339, wf.LastSyncAt)
			_, localUpdate := watcher.statusStore.LastStatus(wf.Name)
			if localUpdate.After(lastUpdate) {
				continue
			}

			j, err := watcher.syncQueue.EnqueueOneByDef(wf.Name)
			if err != nil {
				log.WithFields(log.Fields{
					"remote_last_update": lastUpdate,
					"local_last_update":  localUpdate,
				}).Errorf("create sync job for %s failed: %s", wf.Name, err)
				continue
			}

			log.WithFields(log.Fields{
				"job_id":             j.ID,
				"remote_last_update": lastUpdate,
				"local_last_update":  localUpdate,
			}).Infof("create sync job %s ok", wf.Name)
		}
	}
}
