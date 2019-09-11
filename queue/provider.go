package queue

import (
	"context"
	"sync"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/queue/action"
	"github.com/mylxsw/sync/storage"
)

type ServiceProvider struct{}

func (s *ServiceProvider) Register(app *container.Container) {
	app.MustSingleton(func(cc *container.Container, factory storage.QueueStoreFactory, failedJobStore storage.FailedJobStore) SyncQueue {
		return NewSyncQueue(cc, factory.Queue(storage.QueueFileSync), failedJobStore)
	})
	app.MustSingleton(action.NewActionFactory)
}

func (s *ServiceProvider) Boot(app *glacier.Glacier) {
}

func (s *ServiceProvider) Daemon(ctx context.Context, app *glacier.Glacier) {
	app.MustResolve(func(sq SyncQueue, conf *config.Config) {
		// 注意，worker 是阻塞执行的
		var wg sync.WaitGroup
		wg.Add(conf.FileSyncWorkerNum)
		for i := 0; i < conf.FileSyncWorkerNum; i++ {
			go func(num int) {
				defer wg.Done()
				log.Debugf("sync queue worker [%d] starting ...", num)
				sq.Worker(ctx)
				log.Debugf("sync queue worker [%d] stopped", num)
			}(i)
		}
		wg.Wait()
	})
}
