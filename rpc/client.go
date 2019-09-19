package rpc

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/protocol"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Factory RPC 客户端创建工厂
type Factory interface {
	// SyncClient 创建一个文件同步客户端，使用完成之后不要忘记执行 Close 关闭连接
	SyncClient(endpoint string, token string) (client.FileSyncClient, error)
}

type factory struct{}

// NewFactory 创建一个客户端创建工厂
func NewFactory() Factory {
	return &factory{}
}

// SyncClient 创建一个文件同步客户端
func (factory *factory) SyncClient(endpoint string, token string) (client.FileSyncClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(), grpc.WithPerRPCCredentials(NewAuthAPI(token)))
	if err != nil {
		return nil, errors.Wrap(err, "can't dial to remote rpc server")
	}

	return client.NewFileSyncClient(protocol.NewSyncServiceClient(conn), func() {
		_ = conn.Close()
		log.Debugf("grpc connection closed")
	}), nil
}
