package scheduler

import (
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/go-toolkit/misc"
	"github.com/robfig/cron"
)

type ServiceProvider struct{}

func (s *ServiceProvider) Register(app *container.Container) {
}

func (s *ServiceProvider) Boot(app *glacier.Glacier) {
	app.Crontab(func(cr *cron.Cron, cc *container.Container) error {
		misc.AssertError(cr.AddFunc("@every 3h", wrap(app, ClearJobHistory)))
		return nil
	})
}

func wrap(app *glacier.Glacier, cb interface{}) func() {
	return func() {
		app.MustResolve(cb)
	}
}
