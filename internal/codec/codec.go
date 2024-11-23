package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"gopkg.in/hraban/opus.v2"
)

const MaxPacketSize = 1000
const FramesPerBuffer = Channels * 20 * SampleRate / 1000
const SampleRate = 48000
const Channels = 1

var order = binary.BigEndian

func Write(w io.Writer, sound []float32) error {
	return binary.Write(w, order, sound)
}

func read(r io.Reader, sound []float32) error {
	return binary.Read(r, order, sound)
}

type packet struct {
	DataSize uint16
	Data     [MaxPacketSize]byte
}

type Encoder interface {
	Encode(sound []float32) ([]byte, error)
}

func NewEncoder() (Encoder, error) {
	enc, err := opus.NewEncoder(SampleRate, Channels, opus.AppVoIP)
	if err != nil {
		return nil, err
	}

	return &opusEncoder{
		enc: enc,
		buf: make([]byte, 1024),
	}, nil
}

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

type Decoder interface {
	Decode(data []byte) ([]float32, error)
}

func NewDecoder() (Decoder, error) {
	dec, err := opus.NewDecoder(SampleRate, Channels)
	if err != nil {
		return nil, err
	}

	return &opusDecoder{
		dec: dec,
		buf: make([]byte, 1024),
	}, nil
}

type opusDecoder struct {
	mu  sync.Mutex
	dec *opus.Decoder
	buf []byte
}

func (e *opusDecoder) Decode(data []byte) ([]float32, error) {
	var soundBuf [FramesPerBuffer]float32
	_, err := e.dec.DecodeFloat32(data, soundBuf[:])
	if err != nil {
		return nil, fmt.Errorf("error reading from opus stream: %w", err)
	}

	return soundBuf[:], nil

}
