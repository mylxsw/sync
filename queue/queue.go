package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/queue/job"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/storage"
	"github.com/mylxsw/sync/utils"
)

// SyncQueue 任务同步队列接口
type SyncQueue interface {
	// Enqueue 添加任务到队列
	Enqueue(jobs ...job.FileSyncJob) error
	// Worker 执行任务队列消费
	Worker(ctx context.Context)
	// RunningJobs 获取运行中的任务
	RunningJobs() []storage.JobHistoryItem
}

type syncQueue struct {
	queue       storage.QueueStore
	failedStore storage.FailedJobStore
	statusStore storage.JobStatusStore
	cc          *container.Container

	lock        sync.RWMutex
	runningJobs map[string]*storage.JobHistoryItem
}

// NewSyncQueue 创建一个任务队列
func NewSyncQueue(cc *container.Container, queue storage.QueueStore, failedStore storage.FailedJobStore) SyncQueue {
	sq := syncQueue{queue: queue, failedStore: failedStore, cc: cc, runningJobs: make(map[string]*storage.JobHistoryItem)}

	cc.MustResolve(func(statusStore storage.JobStatusStore) {
		sq.statusStore = statusStore
	})

	return &sq
}

func (sq *syncQueue) AddRunningJob(id string, job *storage.JobHistoryItem) {
	sq.lock.Lock()
	defer sq.lock.Unlock()

	sq.runningJobs[id] = job
}

func (sq *syncQueue) RemoveRunningJob(id string) {
	sq.lock.Lock()
	defer sq.lock.Unlock()

	delete(sq.runningJobs, id)
}

func (sq *syncQueue) Enqueue(jobs ...job.FileSyncJob) error {
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
			defer func() {
				ok <- struct{}{}
			}()

			sq.syncJob()
		}()

		select {
		case <-ctx.Done():
			<-ok
			return
		case <-ok:
		}
	}
}

// syncJob
func (sq *syncQueue) syncJob() {
	var err error
	var historyRecorder func(jobHistory storage.JobHistoryStore)

	defer func() {
		if err2 := recover(); err2 != nil {
			log.Errorf("worker panic and recovered: %s", err2)
			err = fmt.Errorf("worker panic: %s", err2)
		}

		// 记录job执行历史
		if historyRecorder != nil {
			sq.cc.MustResolve(historyRecorder)
		}
	}()

	// 从队列中 pop 一个job
	// 阻塞执行
	var data []byte
	data, err = sq.queue.Dequeue(3 * time.Second)
	if err != nil {
		if err != storage.ErrQueueTimeout {
			log.Errorf("dequeue failed: %s", err)
		}

		return
	}

	// 初始化 job
	j := &job.FileSyncJob{}
	j.Decode(data)

	log.WithFields(log.Fields{
		"job": j,
	}).Debugf("processing job %s [%s]", j.Name, j.ID)

	// 更新任务状态
	if err := sq.statusStore.Update(j.ID, storage.JobStatusRunning); err != nil {
		log.WithFields(log.Fields{
			"job":    j,
			"status": storage.JobStatusRunning,
		}).Errorf("update job status failed: %s", err)
	}

	jhi := storage.JobHistoryItem{
		ID:        utils.UUID(),
		JobID:     j.ID,
		Name:      j.Name,
		Payload:   data,
		Status:    "running",
		CreatedAt: time.Now(),
	}
	sq.AddRunningJob(j.ID, &jhi)

	// 初始化任务执行历史纪录函数
	// 在前面的 defer 中会自动执行该函数
	// 创建数据采集器，用于采集 job 执行过程中的输出
	// 方便记录到执行历史纪录中
	var col = collector.NewCollector(sq.cc.MustGet(&collector.Collectors{}).(*collector.Collectors), j.ID)
	historyRecorder = func(jobHistory storage.JobHistoryStore) {
		defer func() {
			col.Finish()
		}()

		status := "ok"
		if err != nil {
			status = err.Error()
		} else if col.HasError() {
			status = "unstable"
		}

		jhi.Status = status
		jhi.Output = col.Build()

		if err := jobHistory.Record(jhi); err != nil {
			log.WithFields(log.Fields{
				"job": j,
			}).Errorf("record job history failed: %s", err)
		}

		sq.RemoveRunningJob(j.ID)

		// 更新任务状态
		var jobStatus storage.JobStatus
		switch status {
		case "ok":
			jobStatus = storage.JobStatusOK
		case "unstable":
			jobStatus = storage.JobStatusUnstable
		default:
			jobStatus = storage.JobStatusFailed
		}

		if err := sq.statusStore.Update(j.ID, jobStatus); err != nil {
			log.WithFields(log.Fields{
				"job":    j,
				"status": jobStatus,
			}).Errorf("update job status failed: %s", err)
		}
	}

	// 任务执行
	if err = sq.cc.ResolveWithError(func(ctx context.Context, rpcFactory rpc.Factory) error {
		return j.Handle(ctx, rpcFactory, col)
	}); err != nil {
		log.WithFields(log.Fields{
			"job": j,
		}).Errorf("job handle failed: %s", err)

		// 任务执行失败，加入到失败队列暂存
		if err2 := sq.failedStore.Add(j.ID, j.Encode()); err2 != nil {
			log.WithFields(log.Fields{
				"job": j,
			}).Errorf("enqueue FileSyncJob to failed queue failed: %s", err)
		}
	}
}

func (sq *syncQueue) RunningJobs() []storage.JobHistoryItem {
	sq.lock.RLock()
	defer sq.lock.RUnlock()

	items := make([]storage.JobHistoryItem, 0)
	for _, j := range sq.runningJobs {
		items = append(items, *j)
	}

	return items
}
