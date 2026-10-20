package wakeup

import (
	"context"
	"sync"
)

// Broadcaster fans one wake signal out to every subscriber. Each subscriber
// channel holds at most one pending signal, so a burst of Notify calls
// collapses into a single wake.
type Broadcaster struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func New() *Broadcaster {
	return &Broadcaster{
		subs: make(map[chan struct{}]struct{}),
	}
}

func (b *Broadcaster) Notify() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Subscribe's channel is closed once ctx is done.
func (b *Broadcaster) Subscribe(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{}, 1)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		delete(b.subs, ch)
		close(ch)
	}()

	return ch
}
