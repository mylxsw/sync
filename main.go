package main

import (
	"net"
	"sync"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/client"
	"github.com/mylxsw/sync/protocol"
	"github.com/mylxsw/sync/server"
	"google.golang.org/grpc"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		listener, err := net.Listen("tcp", ":8818")
		if err != nil {
			panic(err)
		}

		syncServer := server.NewSyncServer(10240)

		grpcServer := grpc.NewServer()
		protocol.RegisterSyncServiceServer(grpcServer, syncServer)

		if err := grpcServer.Serve(listener); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	// TEST

	conn, err := grpc.Dial("127.0.0.1:8818", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	fs := client.NewFileSync(protocol.NewSyncServiceClient(conn))
	if err := fs.Sync("/data/logs"); err != nil {
		log.Errorf("sync failed: %s", err)
	}

	wg.Wait()
}
