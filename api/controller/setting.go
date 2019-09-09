package controller

import (
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/storage"
)

type SettingController struct {
	cc *container.Container
}

func NewSettingController(cc *container.Container) Controller {
	return &SettingController{cc: cc}
}

func (s *SettingController) Register(router *hades.Router) {
	router.Group("/setting/", func(router *hades.Router) {
		router.Get("/global-sync/", s.GlobalSyncSetting)
		router.Post("/global-sync/", s.UpdateGlobalSyncSetting)
	})
}

func (s *SettingController) GlobalSyncSetting(ctx *hades.WebContext, settingFactory storage.SettingFactory) hades.HTTPResponse {
	settingData, err := settingFactory.Namespace(storage.GlobalNamespace).Get(storage.SyncActionSetting)
	if err != nil {
		if err == storage.ErrNoSuchSetting {
			return ctx.YAML(meta.NewGlobalFileSyncSetting())
		}

		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	setting := meta.GlobalFileSyncSetting{}
	if err := setting.Decode(settingData); err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.YAML(setting)
}

func (s *SettingController) UpdateGlobalSyncSetting(ctx *hades.WebContext, settingFactory storage.SettingFactory) hades.HTTPResponse {
	var setting meta.GlobalFileSyncSetting
	if err := ctx.UnmarshalYAML(&setting); err != nil {
		return ctx.Error(err.Error(), http.StatusUnprocessableEntity)
	}

	if err := settingFactory.Namespace(storage.GlobalNamespace).Update(storage.SyncActionSetting, setting.Encode()); err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(hades.M{})
}
