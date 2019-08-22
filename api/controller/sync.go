package controller

import (
	"fmt"
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/rpc"
)

type SyncController struct {
	cc *container.Container
}

func NewSyncController(cc *container.Container) *SyncController {
	return &SyncController{cc: cc}
}

func (s *SyncController) Register(router *hades.Router) {
	router.Group("/sync", func(router *hades.Router) {
		router.Post("/", s.Sync)
	})
}

func (s *SyncController) Sync(ctx *hades.WebContext, req *hades.Request, rpcFactory rpc.Factory) hades.HTTPResponse {
	syncClient, err := rpcFactory.SyncClient("localhost:8818", "")
	if err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	if err := syncClient.Sync("/var/log"); err != nil {
		return ctx.Error(fmt.Sprintf("sync failed: %s", err), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", nil)
}
