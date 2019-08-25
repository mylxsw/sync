package controller

import (
	"net/http"

	"github.com/mylxsw/asteria/log"
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
	router.Group("/job", func(router *hades.Router) {
		router.Post("/{def}/", j.Sync)
		router.Get("/{id}/", j.Status)
	})
}

// Status 任务执行状态查询
func (s *JobController) Status(ctx *hades.WebContext, req *hades.Request, statusStore storage.JobStatusStore) hades.HTTPResponse {
	jobID := req.PathVar("id")
	if jobID == "" {
		return ctx.Error("invalid job id", http.StatusUnprocessableEntity)
	}

	jobStatus := statusStore.Status(jobID)
	return ctx.API("0000", "ok", hades.M{
		"id":     jobID,
		"status": jobStatus,
	})
}

// Sync 发起文件同步
func (s *JobController) Sync(ctx *hades.WebContext, req *hades.Request, syncQueue queue.SyncQueue, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
	def := req.PathVar("def")
	if def == "" {
		return ctx.Error("invalid def argument", http.StatusUnprocessableEntity)
	}

	definition, err := defStore.Get(def)
	if err != nil {
		if err == storage.ErrNoSuchDefinition {
			return ctx.Error(err.Error(), http.StatusNotFound)
		}

		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	// create file sync job and push to queue
	job := queue.NewFileSyncJob(definition)

	// 记录 job 状态，用于异步查询任务执行状态
	if err := statusStore.Update(job.ID, storage.JobStatusPending); err != nil {
		log.Errorf("record job status failed: %s", err)
	}

	if err := syncQueue.Enqueue(*job); err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", hades.M{
		"id": job.ID,
	})
}
