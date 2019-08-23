package main

import (
	"fmt"
	"os"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/sync/api"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/scheduler"
	"github.com/mylxsw/sync/server"
	"github.com/mylxsw/sync/storage"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var Version string
var GitCommit string

func main() {
	log.DefaultDynamicModuleName(true)
	app := glacier.Create(fmt.Sprintf("%s (%s)", Version, GitCommit))

	app.AddFlags(altsrc.NewInt64Flag(cli.Int64Flag{
		Name:  "file_transfer_buffer_size",
		Usage: "文件传输缓冲区大小",
		Value: 10240,
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "rpc_listen_addr",
		Usage: "GRPC 服务监听地址，用于内部不同的服务实例之间通信",
		Value: ":8818",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "db",
		Usage: "本地数据库存储文件",
		Value: "/data/sync_db",
	}))
	app.AddFlags(altsrc.NewIntFlag(cli.IntFlag{
		Name:  "file_sync_worker_num",
		Usage: "文件同步 Worker 数量",
		Value: 3,
	}))
	app.AddFlags(altsrc.NewInt64Flag(cli.Int64Flag{
		Name:  "job_history_keep_size",
		Usage: "任务执行历史纪录保持数量",
		Value: 20,
	}))

	app.Singleton(func(c *cli.Context) *config.Config {
		return &config.Config{
			FileTransferBufferSize: c.Int64("file_transfer_buffer_size"),
			RPCListenAddr:          c.String("rpc_listen_addr"),
			DB:                     c.String("db"),
			FileSyncWorkerNum:      c.Int("file_sync_worker_num"),
			JobHistoryKeepSize:     c.Int64("job_history_keep_size"),
		}
	})

	app.WithHttpServer(":8819")

	app.Provider(&server.ServiceProvider{})
	app.Provider(&rpc.ServiceProvider{})
	app.Provider(&api.ServiceProvider{})
	app.Provider(&storage.ServiceProvider{})
	app.Provider(&queue.ServiceProvider{})
	app.Provider(&scheduler.ServiceProvider{})

	if err := app.Run(os.Args); err != nil {
		log.Errorf("exit: %s", err)
	}
}
