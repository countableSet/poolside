package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/countableset/poolside/margarita/config"

	"github.com/countableset/poolside/margarita/callbacks"
	"github.com/countableset/poolside/margarita/server"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	config.Load()

	port := config.GetXdsPort()
	cb, sig := callbacks.NewCallbacks()
	ctx := context.Background()
	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)
	srv := xds.NewServer(ctx, snapshotCache, cb)

	go server.RunManagementServer(ctx, srv, port)

	<-sig

	snap := server.DemoData()
	snapshotCache.SetSnapshot(snapshotCache.GetStatusKeys()[0], snap)

	<-done
	log.Print("service stopped, shutting down...")

	ctx.Done()

	log.Print("service exited properly")
}
