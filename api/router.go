package api

import (
	"errors"

	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/api/controller"
	"github.com/mylxsw/sync/config"
)

func routers(cc *container.Container) func(router *hades.Router, mw hades.RequestMiddleware) {
	conf := cc.MustGet(&config.Config{}).(*config.Config)
	return func(router *hades.Router, mw hades.RequestMiddleware) {
		mws := make([]hades.HandlerDecorator, 0)
		mws = append(mws, mw.AccessLog(), mw.CORS("*"), mw.JSONExceptionHandler())
		if conf.APIToken != "" {
			authMiddleware := mw.AuthHandler(func(typ string, credential string) error {
				if typ != "Bearer" {
					return errors.New("invalid auth type, only support Bearer")
				}

				if credential != conf.APIToken {
					return errors.New("token not match")
				}

				return nil
			})

			mws = append(mws, authMiddleware)
		}

		router.Group("/api", func(router *hades.Router) {
			controller.NewWelcomeController(cc).Register(router)
			controller.NewFileSyncController(cc).Register(router)
			controller.NewHistoryController(cc).Register(router)
			controller.NewJobController(cc).Register(router)
			controller.NewSettingController(cc).Register(router)
		}, mws...)
	}
}
