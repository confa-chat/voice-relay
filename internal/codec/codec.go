package codec

import (
	"encoding/binary"
	"io"
	// "gopkg.in/hraban/opus.v2"
)

const MaxPacketSize = 1000
const FramesPerBuffer = Channels * 20 * SampleRate / 1000
const SampleRate = 48000
const Channels = 1

var order = binary.BigEndian

func Write(w io.Writer, sound []float32) error {
	return binary.Write(w, order, sound)
}

func Read(r io.Reader, sound []float32) error {
	return binary.Read(r, order, sound)
}

func WritePacket(w io.Writer, data []byte) error {
	out := [MaxPacketSize]byte{}
	copy(out[:], data)
	return binary.Write(w, order, packet{DataSize: uint16(len(data)), Data: out})
}

func ReadPacket(r io.Reader) ([]byte, error) {
	var packet packet
	err := binary.Read(r, order, &packet)
	if err != nil {
		return nil, err
	}
	return packet.Data[:packet.DataSize], nil
}

type packet struct {
	DataSize uint16
	Data     [MaxPacketSize]byte
}
