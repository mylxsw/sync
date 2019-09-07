package controller

import (
	"fmt"
	"net/http"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/coll"
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/queue/job"
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

	router.Group("/jobs-bulk/", func(router *hades.Router) {
		router.Post("/", j.BulkSync)
	})

	router.Group("/failed-jobs", func(router *hades.Router) {
		router.Get("/", j.FailedJobs)
		router.Put("/{id}/", j.RetryJob)
		router.Delete("/{id}/", j.DeleteFailedJob)
	})
}

type JobStatus struct {
	ID             string `json:"id"`
	DefinitionName string `json:"definition_name,omitempty"`
	Status         string `json:"status"`
}

// Status 任务执行状态查询
// @Summary 查询文件同步任务执行状态
// @Tags Jobs
// @Param id path string true "Job ID"
// @Success 200 {object} controller.JobStatus
// @Router /jobs/{id}/ [get]
func (s *JobController) Status(ctx *hades.WebContext, req *hades.HttpRequest, statusStore storage.JobStatusStore) hades.HTTPResponse {
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

type BulkSyncReq struct {
	Defs []string `json:"defs"`
}

// BulkSync 批量发起文件同步
// @Summary 批量发起文件同步
// @Tags Jobs
// @Param body body controller.BulkSyncReq true "同步定义列表"
// @Success 200 {array} controller.JobStatus
// @Router /jobs-bulk/ [post]
func (s *JobController) BulkSync(ctx *hades.WebContext, syncQueue queue.SyncQueue, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
	var bulkSyncReq BulkSyncReq
	if err := ctx.Unmarshal(&bulkSyncReq); err != nil {
		return ctx.JSONError("invalid request arguments, must be json", http.StatusUnprocessableEntity)
	}

	var definitions []*meta.FileSyncGroup
	for _, df := range bulkSyncReq.Defs {
		def, err := defStore.Get(df)
		if err != nil {
			return ctx.JSONError(fmt.Sprintf("query %s failed: %s", df, err.Error()), http.StatusNotFound)
		}

		definitions = append(definitions, def)
	}

	jobStatuses := make([]JobStatus, 0)
	for _, df := range definitions {
		j := job.NewFileSyncJob(*df)

		// 记录 job 状态，用于异步查询任务执行状态
		if err := statusStore.Update(j.ID, storage.JobStatusPending); err != nil {
			log.Errorf("record job status failed: %s", err)
		}

		if err := syncQueue.Enqueue(*j); err != nil {
			jobStatuses = append(jobStatuses, JobStatus{
				ID:             j.ID,
				DefinitionName: df.Name,
				Status:         string(storage.JobStatusFailed),
			})
			log.Errorf("enqueue job failed: %s", err)
		} else {
			jobStatuses = append(jobStatuses, JobStatus{
				ID:             j.ID,
				DefinitionName: df.Name,
				Status:         string(storage.JobStatusPending),
			})
		}
	}

	return ctx.JSON(jobStatuses)
}

// Sync 发起文件同步
// Parameters:
//     - def 同步定义名称
// @Summary 发起文件同步任务
// @Tags Jobs
// @Param def query string true "同步定义名称"
// @Success 200 {object} controller.JobStatus
// @Router /jobs/ [post]
func (s *JobController) Sync(ctx *hades.WebContext, req *hades.HttpRequest, syncQueue queue.SyncQueue, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
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
	j := job.NewFileSyncJob(*definition)

	// 记录 job 状态，用于异步查询任务执行状态
	if err := statusStore.Update(j.ID, storage.JobStatusPending); err != nil {
		log.Errorf("record job status failed: %s", err)
	}

	if err := syncQueue.Enqueue(*j); err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(JobStatus{
		ID:             j.ID,
		DefinitionName: definition.Name,
		Status:         string(storage.JobStatusPending),
	})
}

// Jobs 返回队列中所有任务
// @Summary 返回队列中所有任务
// @Tags Jobs
// @Success 200 {array} queue.FileSyncJob
// @Router /jobs/ [get]
func (s *JobController) Jobs(ctx *hades.WebContext, req *hades.HttpRequest, queueStoreFactory storage.QueueStoreFactory) hades.HTTPResponse {
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
func (s *JobController) FailedJobs(ctx *hades.WebContext, req *hades.HttpRequest, failedJobStore storage.FailedJobStore) hades.HTTPResponse {
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

func (s *JobController) jobs(jobsRaw [][]byte) ([]job.FileSyncJob, error) {
	var jobs []job.FileSyncJob
	if err := coll.Map(jobsRaw, &jobs, func(raw []byte) job.FileSyncJob {
		var j job.FileSyncJob
		j.Decode(raw)

		return j
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
func (s *JobController) DeleteFailedJob(ctx *hades.WebContext, req *hades.HttpRequest, failedStore storage.FailedJobStore) hades.HTTPResponse {
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

	j := job.FileSyncJob{}
	j.Decode(data)

	if err := failedStore.Delete(j.ID); err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(j)
}

// RetryJob 重试失败的任务
// @Summary 重试失败的任务
// @Tags FailedJobs
// @Param id path string true "要重试的 Job ID"
// @Success 200 {object} queue.FileSyncJob
// @Router /failed-jobs/{id}/ [put]
func (s *JobController) RetryJob(ctx *hades.WebContext, req *hades.HttpRequest, failedStore storage.FailedJobStore, jobQueue queue.SyncQueue) hades.HTTPResponse {
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

	j := job.FileSyncJob{}
	j.Decode(data)

	if err := jobQueue.Enqueue(j); err != nil {
		return ctx.JSONError(fmt.Sprintf("retry job failed: %s", err), http.StatusInternalServerError)
	}

	if err := failedStore.Delete(j.ID); err != nil {
		log.WithFields(log.Fields{
			"job": j,
		}).Errorf("remove failed job failed: %s", err)
	}

	return ctx.JSON(j)
}
