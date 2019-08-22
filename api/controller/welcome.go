package controller

import (
	"github.com/mylxsw/hades"
)

type WelcomeController struct{}

func NewWelcomeController() Controller {
	return &WelcomeController{}
}

func (w *WelcomeController) Register(router *hades.Router) {
	router.Any("/", w.Home)
}

func (w *WelcomeController) Home(ctx *hades.WebContext, req *hades.Request) hades.HTTPResponse {
	return ctx.API("0000", "Hello, world", nil)
}
