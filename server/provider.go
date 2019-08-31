package server

import (
	"context"
	"math"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/graceful"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/protocol"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type ServiceProvider struct{}

func (s *ServiceProvider) Register(app *container.Container) {
	app.MustSingleton(func(conf *config.Config, cc *container.Container) *grpc.Server {
		authFunc := s.authFunc(cc)
		return grpc.NewServer(
			grpc.MaxRecvMsgSize(math.MaxInt32),
			grpc_middleware.WithStreamServerChain(
				grpc_auth.StreamServerInterceptor(authFunc),
				grpc_recovery.StreamServerInterceptor(),
			),
			grpc_middleware.WithUnaryServerChain(
				grpc_auth.UnaryServerInterceptor(authFunc),
				grpc_recovery.UnaryServerInterceptor(),
			),
		)
	})
}

func (s *ServiceProvider) Boot(app *glacier.Glacier) {
	app.MustResolve(func(server *grpc.Server, conf *config.Config) {
		// 注册 GRPC Service
		protocol.RegisterSyncServiceServer(server, NewSyncServer(conf.FileTransferBufferSize))
	})
}

func (s *ServiceProvider) Daemon(ctx context.Context, app *glacier.Glacier) {
	app.MustResolve(func(server *grpc.Server, gf *graceful.Graceful, conf *config.Config) error {
		listener, err := net.Listen("tcp", conf.RPCListenAddr)
		if err != nil {
			return errors.Wrapf(err, "grpc listen to addr %d failed", conf.RPCListenAddr)
		}

		// 平滑关闭 Server
		gf.AddShutdownHandler(func() {
			log.Warning("closing grpc server...")
			server.GracefulStop()
		})

		log.Debugf("grpc server started, listening on %s", conf.RPCListenAddr)
		if err := server.Serve(listener); err != nil {
			return errors.Wrap(err, "grpc server stopped")
		}

		return nil
	})
}

// authFunc 鉴权中间件
func (s *ServiceProvider) authFunc(cc *container.Container) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	}
}
