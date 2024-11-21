package main

import (
	"fmt"
	"io"
	"konfa/voip/internal/codec"

	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
)

func receiveAudio(r io.Reader) error {
	devs, err := portaudio.Devices()
	if err != nil {
		return err
	}

	var dev *portaudio.DeviceInfo
	for _, v := range devs {
		if v.Name == "pipewire" {
			dev = v
			break
		}
	}

	if dev == nil {
		return fmt.Errorf("device not found")
	}

	params := portaudio.LowLatencyParameters(nil, dev)
	params.SampleRate = codec.SampleRate
	params.FramesPerBuffer = codec.FramesPerBuffer
	params.Output.Channels = codec.Channels

	buf := make([]float32, params.FramesPerBuffer)
	stream, err := portaudio.OpenStream(params, buf)
	if err != nil {
		return err
	}
	stream.Start()
	defer stream.Close()

	dec, err := opus.NewDecoder(codec.SampleRate, codec.Channels)
	if err != nil {
		return fmt.Errorf("error creating opus stream: %w", err)
	}

	for {
		data, err := codec.ReadPacket(r)
		if err != nil {
			return fmt.Errorf("error reading packet: %w", err)
		}

		_, err = dec.DecodeFloat32(data, buf)
		if err != nil {
			return fmt.Errorf("error reading from opus stream: %w", err)
		}
		err = stream.Write()
		if err != nil {
			return fmt.Errorf("error writing to stream: %w", err)
		}
	}
}
