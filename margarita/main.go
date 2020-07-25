package main

import (
	"context"

	"github.com/countableset/poolside/margarita/callbacks"
	"github.com/countableset/poolside/margarita/server"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
)

func main() {
	port := uint(8080)
	cb, signal := callbacks.NewCallbacks()
	ctx := context.Background()
	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, nil)
	srv := xds.NewServer(ctx, snapshotCache, cb)

	go server.RunManagementServer(ctx, srv, port)

	<-signal
}
