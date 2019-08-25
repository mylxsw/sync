package storage

import (
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/sync/config"
	"github.com/pkg/errors"
	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
)

type ServiceProvider struct{}

func (s *ServiceProvider) Register(app *container.Container) {
	app.MustSingleton(func(conf *config.Config) (*ledis.DB, error) {
		cfg := lediscfg.NewConfigDefault()
		cfg.DataDir = conf.DB

		conn, err := ledis.Open(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "open ledis database failed")
		}

		db, err := conn.Select(0)
		if err != nil {
			return nil, errors.Wrap(err, "select ledis database failed")
		}

		return db, nil
	})

	app.MustSingleton(NewQueueStoreFactory)

	app.MustSingleton(NewJobHistoryStore)
	app.MustSingleton(NewJobStatusStore)
	app.MustSingleton(NewDefinitionStore)
}

func (s *ServiceProvider) Boot(app *glacier.Glacier) {

}
