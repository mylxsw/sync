package controller

import (
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/hades"
	"github.com/mylxsw/sync/storage"
)

type ErrorController struct {
	cc *container.Container
}

func NewErrorController(cc *container.Container) Controller {
	return &ErrorController{cc: cc}
}

func (e *ErrorController) Register(router *hades.Router) {
	router.Group("/errors/", func(router *hades.Router) {
		router.Get("/", e.Recently)
	})
}

// Recently 返回最近的错误日志
// @Summary 返回最近的错误日志
// @Tags Errors
// @Param limit query int false "返回最近的错误日志数目"
// @Success 200 {array} string
// @Router /errors/ [get]
func (e *ErrorController) Recently(ctx *hades.WebContext, req *hades.HttpRequest, messageFactory storage.MessageFactory) hades.HTTPResponse {
	limit := req.IntInput("limit", 100)
	if limit <= 0 || limit > 1000 {
		return ctx.JSONError("invalid limit argument", http.StatusUnprocessableEntity)
	}

	items, err := messageFactory.MessageStore(storage.MessageErrorType).Recently(limit)
	if err != nil {
		return ctx.JSONError(err.Error(), http.StatusInternalServerError)
	}

	return ctx.JSON(items)
}
