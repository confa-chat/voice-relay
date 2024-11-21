package main

import (
	"errors"
	"fmt"
	"io"
	"konfa/voip/internal/codec"
	"net/http"
	"slices"
	"sync"

	"github.com/royalcat/btrgo/btrchannels"
	"gopkg.in/hraban/opus.v2"
)

func main() {
	vc := newVoiceChannel()

	http.HandleFunc("POST /audio/send", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")

		dec, err := opus.NewDecoder(codec.SampleRate, codec.Channels)
		if err != nil {
			err := fmt.Errorf("error creating opus stream: %w", err)
			println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		soundBuf := make([]float32, codec.FramesPerBuffer)
		for {
			data, err := codec.ReadPacket(r.Body)
			if err != nil {
				if errors.Is(err, io.EOF) {
					w.WriteHeader(http.StatusOK)
					return
				}

				err := fmt.Errorf("error reading from reader: %w", err)
				println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = dec.DecodeFloat32(data, soundBuf)
			if err != nil {
				err := fmt.Errorf("error reading from opus stream: %w", err)
				println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			vc.Send(user, slices.Clone(soundBuf))
		}

	})

	http.HandleFunc("GET /audio/receive", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			panic("expected http.ResponseWriter to be an http.Flusher")
		}

		user := r.URL.Query().Get("user")

		w.Header().Set("Connection", "Keep-Alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Type", "audio/wave")

		flusher.Flush()

		soudChan := vc.Listener(user)
		enc, err := opus.NewEncoder(codec.SampleRate, codec.Channels, opus.AppVoIP)
		if err != nil {
			println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		outBuf := make([]byte, 1024)
		for v := range soudChan.Out() {
			n, err := enc.EncodeFloat32(v, outBuf)
			if err != nil {
				println(err.Error())
				return
			}
			codec.WritePacket(w, outBuf[:n])
			flusher.Flush()
		}
	})

	panic(http.ListenAndServe(":8080", nil))
}

type voiceChannel struct {
	mu     sync.Mutex
	voices map[string]*broadcast[[]float32]
}

func newVoiceChannel() *voiceChannel {
	return &voiceChannel{
		voices: make(map[string]*broadcast[[]float32]),
	}
}

func (vc *voiceChannel) Send(user string, sound []float32) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if _, ok := vc.voices[user]; !ok {
		vc.voices[user] = &broadcast[[]float32]{}
	}

	vc.voices[user].Send(sound)
}

func (vc *voiceChannel) Listener(user string) btrchannels.OutChannel[[]float32] {
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
