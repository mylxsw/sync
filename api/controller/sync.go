package controller

import (
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/queue"
)

type FileSyncController struct {
	cc *container.Container
}

func NewFileSyncController(cc *container.Container) Controller {
	return &FileSyncController{cc: cc}
}

func (s *FileSyncController) Register(router *hades.Router) {
	router.Group("/sync", func(router *hades.Router) {
		router.Post("/", s.Sync)
	})
}

func (s *FileSyncController) Sync(ctx *hades.WebContext, req *hades.Request, syncQueue queue.SyncQueue) hades.HTTPResponse {
	group := client.FileSyncGroup{
		Name:  "sync files",
		From:  "localhost:8818",
		Token: "",
		Files: []client.File{
			{
				Src:  "/var/log",
				Dest: "/tmp/",
				After: []client.SyncAction{
					{
						Action: client.Action{
							Command: "systemctl restart all",
						},
						When: "",
					},
				},
			},
		},
	}

	if len(group.Files) == 0 {
		return ctx.Error("invalid path argument", http.StatusUnprocessableEntity)
	}

	job := queue.NewFileSyncJob(group)
	if err := syncQueue.Enqueue(*job); err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", hades.M{
		"id": job.ID,
	})
}
