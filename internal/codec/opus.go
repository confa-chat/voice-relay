package codec

import (
	"fmt"
	"sync"

	"gopkg.in/hraban/opus.v2"
)

const opusBufferSize = 4096

type opusEncoder struct {
	mu  sync.Mutex
	enc *opus.Encoder
	buf []byte
}

func (e *opusEncoder) Encode(sound []float32) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	n, err := e.enc.EncodeFloat32(sound, e.buf)
	if err != nil {
		return nil, err
	}

	out := make([]byte, n)
	copy(out, e.buf[:n])
	return out, nil
}

type opusDecoder struct {
	mu  sync.Mutex
	dec *opus.Decoder
	buf []byte
}

func (e *opusDecoder) Decode(data []byte) ([]float32, error) {
	var soundBuf [FramesPerBuffer * 10]float32
	n, err := e.dec.DecodeFloat32(data, soundBuf[:])
	if err != nil {
		return nil, fmt.Errorf("decoding data: %w", err)
	}

	return soundBuf[:n], nil

}
