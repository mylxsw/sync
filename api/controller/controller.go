package controller

import (
	"github.com/mylxsw/hades"
)

// Controller is a interface for controller
type Controller interface {
	// Register 用于注册路由以及控制器初始化
	Register(router *hades.Router)
}
