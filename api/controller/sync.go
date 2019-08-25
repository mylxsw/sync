package controller

import (
	"fmt"
	"net/http"

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
		router.Post("/", s.UpdateDefinition)
		router.Get("/{id}/", s.QueryDefinition)
	})
}

func (s *SyncDefinitionController) UpdateDefinition(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
	var syncGroupDef meta.FileSyncGroup
	if err := req.Unmarshal(&syncGroupDef); err != nil {
		return ctx.Error(fmt.Sprintf("parse definition failed: %s", err), http.StatusUnprocessableEntity)
	}

	if err := defStore.Update(syncGroupDef); err != nil {
		return ctx.Error(fmt.Sprintf("update sync group definition failed: %s", err), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", nil)
}

func (s *SyncDefinitionController) QueryDefinition(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
	id := req.PathVar("id")
	if id == "" {
		return ctx.Error("invalid argument id", http.StatusUnprocessableEntity)
	}

	def, err := defStore.Get(id)
	if err != nil {
		if err == storage.ErrNoSuchDefinition {
			return ctx.Error(err.Error(), http.StatusNotFound)
		}

		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", hades.M{
		"definition": def,
	})
}

func (s *SyncDefinitionController) AllDefinitions(ctx *hades.WebContext, req *hades.Request, defStore storage.DefinitionStore) hades.HTTPResponse {
	defs, err := defStore.All()
	if err != nil {
		return ctx.Error(err.Error(), http.StatusInternalServerError)
	}

	return ctx.API("0000", "ok", hades.M{
		"definitions": defs,
	})
}
