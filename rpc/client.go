package rpc

import (
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/protocol"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Factory RPC 客户端创建工厂
type Factory interface {
	// SyncClient 创建一个文件同步客户端
	SyncClient(endpoint string, token string) (client.FileSync, error)
}

type factory struct{}

// NewFactory 创建一个客户端创建工厂
func NewFactory() Factory {
	return &factory{}
}

// SyncClient 创建一个文件同步客户端
func (factory *factory) SyncClient(endpoint string, token string) (client.FileSync, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(), grpc.WithPerRPCCredentials(NewAuthAPI(token)))
	if err != nil {
		return nil, errors.Wrap(err, "can't dial to remote rpc server")
	}

	return client.NewFileSync(protocol.NewSyncServiceClient(conn)), nil
}
