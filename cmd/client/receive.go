package main

import (
	"fmt"
	"io"
	"konfa/voip/internal/codec"

	"github.com/gordonklaus/portaudio"
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

	for {
		err := codec.Read(r, buf)
		if err != nil {
			return fmt.Errorf("error reading from reader: %w", err)
		}
		err = stream.Write()
		if err != nil {
			return fmt.Errorf("error writing to stream: %w", err)
		}
	}
}
