package api

import (
	"github.com/gorilla/mux"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	_ "github.com/mylxsw/sync/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Sync API
// @version 1.0
// @description 文件同步服务

// @contact.name mylxsw
// @contact.url https://github.com/mylxsw/sync
// @contact.email mylxsw@aicode.cc

// @license.name MIT
// @license.url https://raw.githubusercontent.com/mylxsw/sync/master/LICENSE

// @host localhost:8819
// @BasePath /api
type ServiceProvider struct{}

func (s ServiceProvider) Register(app *container.Container) {}

func (s ServiceProvider) Boot(app *glacier.Glacier) {
	app.WebAppRouter(routers(app.Container()))
	app.WebAppMuxRouter(func(router *mux.Router) {
		router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	})
}
