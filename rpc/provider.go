package rpc

import (
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
)

type ServiceProvider struct{}

func (s *ServiceProvider) Register(app *container.Container) {
	app.MustSingleton(NewFactory)
}

func (s *ServiceProvider) Boot(app *glacier.Glacier) {}
