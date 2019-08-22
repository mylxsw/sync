package api

import (
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/api/controller"
)

func routers(cc *container.Container) func (router *hades.Router, mw hades.RequestMiddleware) {
	return func(router *hades.Router, mw hades.RequestMiddleware) {
		router.Group("/", func(router *hades.Router) {
			controller.NewWelcomeController().Register(router)
			controller.NewFileSyncController(cc).Register(router)

		}, mw.AccessLog(), mw.JSONExceptionHandler(), mw.CORS("*"))
	}
}
