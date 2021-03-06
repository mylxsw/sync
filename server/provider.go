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
	"github.com/mylxsw/sync/storage"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	app.MustResolve(func(server *grpc.Server, conf *config.Config, statusStore storage.JobStatusStore) {
		// 注册 GRPC Service
		protocol.RegisterSyncServiceServer(server, NewSyncServer(conf.FileTransferBufferSize, statusStore))
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
	var conf = cc.MustGet(&config.Config{}).(*config.Config)
	return func(ctx context.Context) (context.Context, error) {
		if conf.RPCToken != "" {
			meta, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return ctx, status.Errorf(codes.Unauthenticated, "auth failed: token required")
			}

			token := meta.Get("token")
			if len(token) != 1 {
				return ctx, status.Errorf(codes.Unauthenticated, "auth failed: invalid token")
			}

			if conf.RPCToken != token[0] {
				return ctx, status.Errorf(codes.Unauthenticated, "auth failed: token not match")
			}
		}

		return ctx, nil
	}
}
