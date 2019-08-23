package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/storage"
)

// SyncQueue 任务同步队列接口
type SyncQueue interface {
	// Enqueue 添加任务到队列
	Enqueue(jobs ...FileSyncJob) error
	// Worker 执行任务队列消费
	Worker(ctx context.Context)
}

type syncQueue struct {
	queue       storage.Queue
	failedQueue storage.Queue
	cc          *container.Container
}

// NewSyncQueue 创建一个任务队列
func NewSyncQueue(cc *container.Container, queue storage.Queue, failedQueue storage.Queue) SyncQueue {
	return &syncQueue{queue: queue, failedQueue: failedQueue, cc: cc}
}

func (sq *syncQueue) Enqueue(jobs ...FileSyncJob) error {
	for _, j := range jobs {
		log.WithFields(log.Fields{
			"job": j,
		}).Debugf("enqueue a new job %s [%s]", j.Name, j.ID)

		if err := sq.queue.Enqueue(j.Encode()); err != nil {
			return err
		}
	}

	return nil
}

func (sq *syncQueue) Worker(ctx context.Context) {
	ok := make(chan struct{})
	defer close(ok)

	for {
		go func() {
			var err error
			var historyRecorder func(jobHistory storage.JobHistory)
			var col = collector.NewCollector()

			defer func() {
				ok <- struct{}{}
				if err2 := recover(); err2 != nil {
					log.Errorf("worker panic and recovered: %s", err)
					err = fmt.Errorf("worker panic: %s", err)
				}

				if historyRecorder != nil {
					sq.cc.MustResolve(historyRecorder)
				}
			}()

			var data []byte
			data, err = sq.queue.Dequeue(2 * time.Second)
			if err != nil {
				if err != storage.ErrQueueTimeout {
					log.Errorf("dequeue failed: %s", err)
				}

				return
			}

			job := &FileSyncJob{}
			job.Decode(data)

			log.WithFields(log.Fields{
				"job": job,
			}).Debugf("processing job %s [%s]", job.Name, job.ID)

			historyRecorder = func(jobHistory storage.JobHistory) {
				status := "ok"
				if err != nil {
					status = err.Error()
				}

				if err := jobHistory.Record(job.Name, job.ID, data, status, col.Build()); err != nil {
					log.WithFields(log.Fields{
						"job": job,
					}).Errorf("record job history failed: %s", err)
				}
			}

			if err = sq.cc.ResolveWithError(func(ctx context.Context, rpcFactory rpc.Factory) error {
				return job.Handle(ctx, rpcFactory, col)
			}); err != nil {
				log.WithFields(log.Fields{
					"job": job,
				}).Errorf("job handle failed: %s", err)

				if err2 := sq.failedQueue.Enqueue(job.Encode()); err2 != nil {
					log.WithFields(log.Fields{
						"job": job,
					}).Errorf("enqueue FileSyncJob to failed queue failed: %s", err)
				}
			}
		}()

		select {
		case <-ctx.Done():
			<-ok
			return
		case <-ok:
		}
	}
}
