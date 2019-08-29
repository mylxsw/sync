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

// Home 欢迎页面
// @Summary 欢迎页面
// @Accept  json
// @Produce  json
// @Success 200 {string} string
// @Router /accounts/{id} [get]
func (w *WelcomeController) Home(ctx *hades.WebContext, req *hades.Request) string {
	return "Hello, world"
}
