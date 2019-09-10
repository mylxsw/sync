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

// GlobalSyncSetting 全局同步配置
// @Summary 全局同步配置
// @Tags Setting
// @Param format query string false "输出格式：yaml/json"
// @Success 200 {object} meta.GlobalFileSyncSetting
// @Router /setting/global-sync/ [get]
func (s *SettingController) GlobalSyncSetting(ctx *hades.WebContext, settingFactory storage.SettingFactory) hades.HTTPResponse {
	resFormat := ctx.InputWithDefault("format", "json")
	if resFormat != "json" && resFormat != "yaml" {
		return ctx.JSONError("invalid format, only support json/yaml", http.StatusUnprocessableEntity)
	}

	settingData, err := settingFactory.Namespace(storage.GlobalNamespace).Get(storage.SyncActionSetting)
	if err != nil {
		if err == storage.ErrNoSuchSetting {
			if resFormat == "json" {
				return ctx.JSON(meta.NewGlobalFileSyncSetting())
			} else {
				return ctx.YAML(meta.NewGlobalFileSyncSetting())
			}
		}

		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	setting := meta.GlobalFileSyncSetting{}
	if err := setting.Decode(settingData); err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	if resFormat == "json" {
		return ctx.JSON(setting)
	}

	return ctx.YAML(setting)
}

// UpdateGlobalSyncSetting 更新全局同步配置
// @Summary 更新全局同步配置
// @Tags Setting
// @Param def body meta.GlobalFileSyncSetting true "全局同步定义"
// @Router /setting/global-sync/ [post]
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
