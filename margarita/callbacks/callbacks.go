package callbacks

import (
	"context"
	"log"
	"sync"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

type Callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
}

func NewCallbacks() (*Callbacks, chan struct{}) {
	signal := make(chan struct{})
	return &Callbacks{
		signal:   signal,
		fetches:  0,
		requests: 0,
	}, signal
}
func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	// log.WithFields(log.Fields{"fetches": cb.fetches, "requests": cb.requests}).Info("cb.Report()  callbacks")
}
func (cb *Callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Printf("OnStreamOpen %d open for %s", id, typ)
	return nil
}
func (cb *Callbacks) OnStreamClosed(id int64) {
	log.Printf("OnStreamClosed %d closed", id)
}
func (cb *Callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest) error {
	log.Printf("OnStreamRequest")
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.requests++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *Callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {
	log.Printf("OnStreamResponse...")
	cb.Report()
}
func (cb *Callbacks) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	log.Printf("OnFetchRequest...")
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.fetches++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *Callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
