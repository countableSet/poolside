package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/countableset/poolside/margarita/api"
	"github.com/countableset/poolside/margarita/config"
	"github.com/countableset/poolside/margarita/server"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	config.Load()

	port := config.GetXdsPort()
	ctx := context.Background()
	snapshotCache := cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil)
	cb := &testv3.Callbacks{Debug: true}
	srv := serverv3.NewServer(ctx, snapshotCache, cb)

	go api.RunAPIServer()
	go server.RunManagementServer(ctx, srv, port)
	go server.ListenForConfigurationChanges(snapshotCache)

	<-done
	log.Print("service stopped, shutting down...")

	ctx.Done()

	log.Print("service exited properly")
}
