package codec

import (
	"encoding/binary"
	"io"
	// "gopkg.in/hraban/opus.v2"
)

const FramesPerBuffer = 128
const SampleRate = 44100
const Channels = 1

var order = binary.BigEndian

func Write(w io.Writer, sound []float32) error {
	return binary.Write(w, order, sound)
}

func Read(r io.Reader, sound []float32) error {
	return binary.Read(r, order, sound)
}
