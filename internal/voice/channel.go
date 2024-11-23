package voice

import (
	"sync"

	"github.com/royalcat/btrgo/btrchannels"
)

type Channel struct {
	mu     sync.Mutex
	voices map[string]*broadcast[[]float32]
}

func NewChannel() *Channel {
	return &Channel{
		voices: make(map[string]*broadcast[[]float32]),
	}
}

func (vc *Channel) Users() []string {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	out := make([]string, 0, len(vc.voices))
	for u := range vc.voices {
		out = append(out, u)
	}
	return out
}

func (vc *Channel) Send(user string, sound []float32) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if _, ok := vc.voices[user]; !ok {
		vc.voices[user] = &broadcast[[]float32]{}
	}

	vc.voices[user].Send(sound)
}

func (vc *Channel) Listener(user string) btrchannels.OutChannel[[]float32] {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if _, ok := vc.voices[user]; !ok {
		vc.voices[user] = &broadcast[[]float32]{}
	}

	return vc.voices[user].ReceiveChan()
}

type broadcast[T any] struct {
	mu        sync.Mutex
	callbacks []func(T)
}

func (b *broadcast[T]) Send(v T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, cb := range b.callbacks {
		cb(v)
	}
}

func (b *broadcast[T]) ReceiveChan() btrchannels.OutChannel[T] {
	ch := btrchannels.NewRingChannel[T](10)
	b.callbacks = append(b.callbacks, func(v T) {
		ch.In() <- v
	})
	return ch
}
