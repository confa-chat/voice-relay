package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"gopkg.in/hraban/opus.v2"
)

type Codec int

const (
	CodecUnknown Codec = iota
	CodecOpus
	CodecPCMf32
)

const MaxPacketSize = 1000
const FramesPerBuffer = Channels * 20 * SampleRate / 1000
const SampleRate = 48000
const Channels = 1

var order = binary.LittleEndian

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

type Decoder interface {
	Decode(data []byte) ([]float32, error)
}

func NewEncoder(codec Codec) (Encoder, error) {
	switch codec {
	case CodecOpus:
		enc, err := opus.NewEncoder(SampleRate, Channels, opus.AppVoIP)
		if err != nil {
			return nil, err
		}

		return &opusEncoder{
			enc: enc,
			buf: make([]byte, opusBufferSize),
		}, nil
	case CodecPCMf32:
		return &pcmF32Encoder{}, nil
	default:
		return nil, fmt.Errorf("unknown codec: %d", codec)

	}
}

func NewDecoder(codec Codec) (Decoder, error) {
	switch codec {
	case CodecOpus:
		dec, err := opus.NewDecoder(SampleRate, Channels)
		if err != nil {
			return nil, err
		}
		return &opusDecoder{
			dec: dec,
			buf: make([]byte, opusBufferSize),
		}, nil
	case CodecPCMf32:
		return &pcmF32Decoder{}, nil
	default:
		return nil, fmt.Errorf("unknown codec: %d", codec)
	}
}
