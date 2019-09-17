package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mylxsw/asteria/formatter"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/glacier"
	"github.com/mylxsw/sync/api"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/queue"
	"github.com/mylxsw/sync/rpc"
	"github.com/mylxsw/sync/scheduler"
	"github.com/mylxsw/sync/server"
	"github.com/mylxsw/sync/storage"
	"github.com/mylxsw/sync/utils"
	"github.com/urfave/cli"
	"github.com/urfave/cli/altsrc"
)

var Version string
var GitCommit string

func main() {
	log.DefaultDynamicModuleName(true)
	app := glacier.Create(fmt.Sprintf("%s (%s)", Version, GitCommit[:8]))

	app.AddFlags(altsrc.NewInt64Flag(cli.Int64Flag{
		Name:  "file_transfer_buffer_size",
		Usage: "file transfer buffer size for rpc stream",
		Value: 10240,
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "rpc_listen_addr",
		Usage: "GRPC server listen addr for internal server communication",
		Value: ":8818",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "rpc_token",
		Usage: "GRPC access token",
		Value: "",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "api_token",
		Usage: "API Token for api access control",
		Value: "",
	}))
	app.AddFlags(altsrc.NewStringFlag(cli.StringFlag{
		Name:  "db",
		Usage: "local database storage path",
		Value: "/data/sync_db",
	}))
	app.AddFlags(altsrc.NewIntFlag(cli.IntFlag{
		Name:  "file_sync_worker_num",
		Usage: "worker count for file sync",
		Value: 3,
	}))
	app.AddFlags(altsrc.NewInt64Flag(cli.Int64Flag{
		Name:  "job_history_keep_size",
		Usage: "job history's count to keep",
		Value: 100,
	}))
	app.AddFlags(altsrc.NewBoolTFlag(cli.BoolTFlag{
		Name:  "console_color",
		Usage: "log colorful for console",
	}))
	app.AddFlags(altsrc.NewBoolFlag(cli.BoolFlag{
		Name:  "use_local_dashboard",
		Usage: "whether using local dashboard, this is used when development",
	}))
	app.AddFlags(altsrc.NewStringSliceFlag(cli.StringSliceFlag{
		Name:  "allow_files",
		Usage: "limit files or directories can be sync",
		Value: &cli.StringSlice{"/data",},
	}))

	app.BeforeInitialize(func(c *cli.Context) error {
		log.DefaultLogFormatter(formatter.NewDefaultFormatter(c.Bool("console_color")))
		return nil
	})

	app.Singleton(func(c *cli.Context) *config.Config {
		return &config.Config{
			FileTransferBufferSize: c.Int64("file_transfer_buffer_size"),
			RPCListenAddr:          c.String("rpc_listen_addr"),
			DB:                     c.String("db"),
			FileSyncWorkerNum:      c.Int("file_sync_worker_num"),
			JobHistoryKeepSize:     c.Int64("job_history_keep_size"),
			RPCToken:               c.String("rpc_token"),
			APIToken:               c.String("api_token"),
			UseLocalDashboard:      c.Bool("use_local_dashboard"),
			AllowFiles:             utils.StringArrayUnique(c.StringSlice("allow_files")),
			CommandTimeout:         60 * time.Second,
		}
	})

	app.WithHttpServer(":8819")

	app.Provider(&server.ServiceProvider{})
	app.Provider(&rpc.ServiceProvider{})
	app.Provider(&api.ServiceProvider{})
	app.Provider(&storage.ServiceProvider{})
	app.Provider(&queue.ServiceProvider{})
	app.Provider(&scheduler.ServiceProvider{})
	app.Provider(&collector.ServiceProvider{})

	app.Main(func(conf *config.Config) {
		log.WithFields(log.Fields{
			"config": conf,
		}).Debug("configuration")
	})

	if err := app.Run(os.Args); err != nil {
		log.Errorf("exit: %s", err)
	}
}
