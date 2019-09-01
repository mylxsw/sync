package controller

import (
	"fmt"
	"net/http"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/meta"
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
func (s *SyncDefinitionController) UpdateDefinitions(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
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

// QueryDefinition get a definition by id
// @Summary 查询单个文件同步定义
// @Tags Sync
// @Param id query string true "定义名称"
// @Param format query string false "输出格式：yaml/json"
// @Success 200 {array} meta.FileSyncGroup
// @Router /sync/{name}/ [get]
func (s *SyncDefinitionController) QueryDefinition(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
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
func (s *SyncDefinitionController) AllDefinitions(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
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
