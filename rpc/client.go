package rpc

import (
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/protocol"
	"google.golang.org/grpc"
)

type Factory interface {
	SyncClient(endpoint string, token string) (client.FileSync, error)
}

type factory struct{}

func NewFactory() Factory {
	return &factory{}
}

func (factory *factory) SyncClient(endpoint string, token string) (client.FileSync, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(), grpc.WithPerRPCCredentials(NewAuthAPI(token)))
	if err != nil {
		return nil, err
	}

	return client.NewFileSync(protocol.NewSyncServiceClient(conn)), nil
}
