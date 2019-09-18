package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/storage"
)

type SyncDefinitionController struct {
	cc *container.Container
}

func NewFileSyncController(cc *container.Container) Controller {
	return &SyncDefinitionController{cc: cc}
}

func (s *SyncDefinitionController) Register(router *hades.Router) {
	router.Group("/sync", func(router *hades.Router) {
		router.Get("/", s.AllDefinitions)
		router.Post("/", s.UpdateDefinitions)
		router.Get("/{name}/", s.QueryDefinition)
		router.Delete("/{name}/", s.DeleteDefinition)
	})

	router.Group("/sync-bulk/", func(router *hades.Router) {
		router.Delete("/", s.BulkDelete)
	})

	router.Group("/sync-stat/", func(router *hades.Router) {
		router.Get("/", s.AllDefinitionStatus)
	})
}

// UpdateDefinitions update sync definitions
// @Summary 更新文件同步定义
// @Tags Sync
// @Accept json
// @Produce json
// @Param def body []meta.FileSyncGroup true "文件同步定义"
// @Success 200 {array} meta.FileSyncGroup
// @Router /sync/ [post]
func (s *SyncDefinitionController) UpdateDefinitions(ctx *hades.WebContext, req *hades.HttpRequest, defStore storage.DefinitionStore, conf *config.Config) hades.HTTPResponse {
	var syncGroupDefs []meta.FileSyncGroup
	if req.ContentType() == "application/yaml" {
		if err := req.UnmarshalYAML(&syncGroupDefs); err != nil {
			return ctx.JSONError(fmt.Sprintf("parse definition failed: %s", err), http.StatusUnprocessableEntity)
		}
	} else {
		if err := req.Unmarshal(&syncGroupDefs); err != nil {
			return ctx.JSONError(fmt.Sprintf("parse definition failed: %s", err), http.StatusUnprocessableEntity)
		}
	}

	for _, def := range syncGroupDefs {
		for _, f := range def.Files {
			if !conf.Allow(f.Dest) {
				return ctx.JSONError(fmt.Sprintf("security: dest for %s not allowed to sync: %s", def.Name, f.Dest), http.StatusUnprocessableEntity)
			}
		}
	}

	var results []meta.FileSyncGroup
	for _, syncGroupDef := range syncGroupDefs {
		if err := defStore.Update(syncGroupDef); err != nil {
			return ctx.JSONError(fmt.Sprintf("update sync group definition failed: %s", err), http.StatusInternalServerError)
		}

		res, err := defStore.Get(syncGroupDef.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"original": syncGroupDef,
			}).Errorf("retrieve sync definition failed: %s", err)
		} else {
			results = append(results, *res)
		}
	}

	return ctx.JSON(results)
}

type BulkDeleteReq struct {
	Names []string `json:"names"`
}

// BulkDelete 批量删除同步定义
// @Summary 批量删除同步定义
// @Tags Sync
// @Param body body controller.BulkDeleteReq true "定义名称列表"
// @Success 200 {string} string
// @Router /sync-bulk/ [delete]
func (s *SyncDefinitionController) BulkDelete(ctx *hades.WebContext, syncQueue queue.SyncQueue, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
	var bulkDeleteReq BulkDeleteReq
	if err := ctx.Unmarshal(&bulkDeleteReq); err != nil {
		return ctx.JSONError("invalid request arguments, must be json", http.StatusUnprocessableEntity)
	}

	for _, name := range bulkDeleteReq.Names {
		if err := defStore.Delete(name); err != nil {
			return ctx.JSONError(err.Error(), http.StatusInternalServerError)
		}
	}

	return ctx.JSON(hades.M{})
}

// DeleteDefinition delete a definition by name
// @Summary 删除单个文件同步定义
// @Tags Sync
// @Param name path string true "定义名称"
// @Success 200 {string} string
// @Router /sync/{name}/ [delete]
func (s *SyncDefinitionController) DeleteDefinition(ctx *hades.WebContext, req *hades.HttpRequest, defStore storage.DefinitionStore) hades.HTTPResponse {
	name := req.PathVar("name")
	if name == "" {
		return ctx.JSONError("invalid argument name", http.StatusUnprocessableEntity)
	}

	if err := defStore.Delete(name); err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(hades.M{})
}

// QueryDefinition get a definition by id
// @Summary 查询单个文件同步定义
// @Tags Sync
// @Param name path string true "定义名称"
// @Param format query string false "输出格式：yaml/json"
// @Success 200 {array} meta.FileSyncGroup
// @Router /sync/{name}/ [get]
func (s *SyncDefinitionController) QueryDefinition(ctx *hades.WebContext, req *hades.HttpRequest, defStore storage.DefinitionStore) hades.HTTPResponse {
	name := req.PathVar("name")
	if name == "" {
		return ctx.JSONError("invalid argument name", http.StatusUnprocessableEntity)
	}

	resFormat := req.InputWithDefault("format", "json")
	if resFormat != "json" && resFormat != "yaml" {
		return ctx.JSONError("invalid format, only support json/yaml", http.StatusUnprocessableEntity)
	}

	def, err := defStore.Get(name)
	if err != nil {
		if err == storage.ErrNoSuchDefinition {
			return ctx.JSONError(err.Error(), http.StatusNotFound)
		}

		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	if resFormat == "json" {
		return ctx.JSON([]*meta.FileSyncGroup{def})
	} else {
		return ctx.YAML([]*meta.FileSyncGroup{def})
	}
}

// AllDefinitions return all definitions
// @Summary 查询所有文件同步定义
// @Tags Sync
// @Param format query string false "输出格式：yaml/json"
// @Success 200 {array} meta.FileSyncGroup
// @Router /sync/ [get]
func (s *SyncDefinitionController) AllDefinitions(ctx *hades.WebContext, req *hades.HttpRequest, defStore storage.DefinitionStore, settingFactory storage.SettingFactory) hades.HTTPResponse {
	resFormat := req.InputWithDefault("format", "json")
	if resFormat != "json" && resFormat != "yaml" {
		return ctx.JSONError("invalid format, only support json/yaml", http.StatusUnprocessableEntity)
	}

	defs, err := defStore.All()
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	if resFormat == "json" {
		return ctx.JSON(defs)
	} else {
		return ctx.YAML(defs)
	}
}

type DefinitionStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AllDefinitionStatus return all definition's status
// @Summary 查询所有文件同步定义的状态
// @Tags SyncStatus
// @Success 200 {array} controller.DefinitionStatus
// @Router /sync-stat/ [get]
func (s *SyncDefinitionController) AllDefinitionStatus(ctx *hades.WebContext, defStore storage.DefinitionStore, statusStore storage.JobStatusStore) hades.HTTPResponse {
	defs, err := defStore.All()
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	var definitionStatuses = make([]DefinitionStatus, 0)
	for _, df := range defs {
		stat, lastUpdate := statusStore.LastStatus(df.Name)
		definitionStatuses = append(definitionStatuses, DefinitionStatus{
			Name:      df.Name,
			Status:    string(stat),
			UpdatedAt: lastUpdate,
		})
	}

	return ctx.JSON(definitionStatuses)
}
