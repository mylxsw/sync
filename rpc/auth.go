package rpc

import (
	"context"
)

// AuthAPI 权限插件
type AuthAPI struct {
	token string
}

// NewAuthAPI 创建一个Auth API
func NewAuthAPI(token string) *AuthAPI {
	return &AuthAPI{token: token}
}

func (a *AuthAPI) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"token": a.token,
	}, nil
}

func (a *AuthAPI) RequireTransportSecurity() bool {
	return false
}
