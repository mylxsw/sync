package controller

import (
	"fmt"
	"net/http"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/coll"
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/storage"
)

type JobController struct {
	cc *container.Container
}

func NewJobController(cc *container.Container) Controller {
	return &JobController{cc: cc}
}

func (j *JobController) Register(router *hades.Router) {
	router.Group("/jobs", func(router *hades.Router) {
		router.Post("/", j.Sync)
		router.Get("/", j.Jobs)
		router.Get("/{id}/", j.Status)
	})

	router.Group("/failed-jobs", func(router *hades.Router) {
		router.Get("/", j.FailedJobs)
		router.Put("/{id}/", j.RetryJob)
		router.Delete("/{id}/", j.DeleteFailedJob)
	})
}

type JobStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Status 任务执行状态查询
// @Summary 查询文件同步任务执行状态
// @Tags Jobs
// @Param id path string true "Job ID"
// @Success 200 {object} controller.JobStatus
// @Router /jobs/{id}/ [get]
func (s *JobController) Status(ctx *hades.WebContext, req *hades.Request, statusStore storage.JobStatusStore) hades.HTTPResponse {
	jobID := req.PathVar("id")
	if jobID == "" {
		return ctx.JSONError("invalid job id", http.StatusUnprocessableEntity)
	}

	jobStatus := statusStore.Status(jobID)
	return ctx.JSON(JobStatus{
		ID:     jobID,
		Status: string(jobStatus),
	})
}

// Sync 发起文件同步
// Parameters:
//     - def 同步定义名称
// @Summary 发起文件同步任务
// @Tags Jobs
// @Param def query string true "同步定义名称"
// @Success 200 {object} controller.JobStatus
// @Router /jobs/ [post]
func (s *JobController) Sync(ctx *hades.WebContext, req *hades.Request, syncQueue queue.SyncQueue, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
	def := req.Input("def")
	if def == "" {
		return ctx.JSONError("invalid def argument", http.StatusUnprocessableEntity)
	}

	definition, err := defStore.Get(def)
	if err != nil {
		if err == storage.ErrNoSuchDefinition {
			return ctx.JSONError(err.Error(), http.StatusNotFound)
		}

		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	// create file sync job and push to queue
	job := queue.NewFileSyncJob(*definition)

	// 记录 job 状态，用于异步查询任务执行状态
	if err := statusStore.Update(job.ID, storage.JobStatusPending); err != nil {
		log.Errorf("record job status failed: %s", err)
	}

	if err := syncQueue.Enqueue(*job); err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(JobStatus{
		ID:     job.ID,
		Status: string(storage.JobStatusPending),
	})
}

// Jobs 返回队列中所有任务
// @Summary 返回队列中所有任务
// @Tags Jobs
// @Success 200 {array} queue.FileSyncJob
// @Router /jobs/ [get]
func (s *JobController) Jobs(ctx *hades.WebContext, req *hades.Request, queueStoreFactory storage.QueueStoreFactory) hades.HTTPResponse {
	qs := queueStoreFactory.Queue("file-sync")
	jobsRaw, err := qs.All()
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	jobs, err := s.jobs(jobsRaw)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(jobs)
}

// FailedJobs 返回失败的所有任务
// @Summary 返回失败的所有任务
// @Tags FailedJobs
// @Success 200 {array} queue.FileSyncJob
// @Router /failed-jobs/ [get]
func (s *JobController) FailedJobs(ctx *hades.WebContext, req *hades.Request, failedJobStore storage.FailedJobStore) hades.HTTPResponse {
	jobsRaw, err := failedJobStore.All()
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	jobs, err := s.jobs(jobsRaw)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(jobs)
}

func (s *JobController) jobs(jobsRaw [][]byte) ([]queue.FileSyncJob, error) {
	var jobs []queue.FileSyncJob
	if err := coll.Map(jobsRaw, &jobs, func(raw []byte) queue.FileSyncJob {
		var job queue.FileSyncJob
		job.Decode(raw)

		return job
	}); err != nil {
		return nil, err
	}

	return jobs, nil
}

// DeleteFailedJob 删除失败的任务
// @Summary 删除失败的任务
// @Tags FailedJobs
// @Param id path string true "删除失败的 Job ID"
// @Success 200 {object} queue.FileSyncJob
// @Router /failed-jobs/{id}/ [delete]
func (s *JobController) DeleteFailedJob(ctx *hades.WebContext, req *hades.Request, failedStore storage.FailedJobStore) hades.HTTPResponse {
	id := req.PathVar("id")
	if id == "" {
		return ctx.JSONError("id argument required", http.StatusUnprocessableEntity)
	}

	data, err := failedStore.Get(id)
	if err != nil {
		if err == storage.ErrNoSuchJob {
			return ctx.JSONError("no such job in queue", http.StatusNotFound)
		}

		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	job := queue.FileSyncJob{}
	job.Decode(data)

	if err := failedStore.Delete(job.ID); err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(job)
}

// RetryJob 重试失败的任务
// @Summary 重试失败的任务
// @Tags FailedJobs
// @Param id path string true "要重试的 Job ID"
// @Success 200 {object} queue.FileSyncJob
// @Router /failed-jobs/{id}/ [put]
func (s *JobController) RetryJob(ctx *hades.WebContext, req *hades.Request, failedStore storage.FailedJobStore, jobQueue queue.SyncQueue) hades.HTTPResponse {
	id := req.PathVar("id")
	if id == "" {
		return ctx.JSONError("id argument required", http.StatusUnprocessableEntity)
	}

	data, err := failedStore.Get(id)
	if err != nil {
		if err == storage.ErrNoSuchJob {
			return ctx.JSONError("no such job in queue", http.StatusNotFound)
		}

		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	job := queue.FileSyncJob{}
	job.Decode(data)

	if err := jobQueue.Enqueue(job); err != nil {
		return ctx.JSONError(fmt.Sprintf("retry job failed: %s", err), http.StatusInternalServerError)
	}

	if err := failedStore.Delete(job.ID); err != nil {
		log.WithFields(log.Fields{
			"job": job,
		}).Errorf("remove failed job failed: %s", err)
	}

	return ctx.JSON(job)
}
