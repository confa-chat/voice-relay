package main

import (
	"io"
	"konfa/voip/internal/codec"

	"github.com/gordonklaus/portaudio"
)

func sendAudio(w io.Writer) error {
	devs, err := portaudio.Devices()
	if err != nil {
		return err
	}

	var dev *portaudio.DeviceInfo
	for _, d := range devs {
		if d.Name == "pipewire" {
			dev = d
			break
		}
	}
	println(dev.Name)

	params := portaudio.LowLatencyParameters(dev, nil)
	params.SampleRate = codec.SampleRate
	params.FramesPerBuffer = codec.FramesPerBuffer
	params.Input.Channels = codec.Channels

	buf := make([]float32, params.FramesPerBuffer)
	stream, err := portaudio.OpenStream(params, buf)
	if err != nil {
		return err
	}
	stream.Start()
	defer stream.Close()

	for {
		err := stream.Read()
		if err != nil {
			return err
		}
		err = codec.Write(w, buf)
		if err != nil {
			return err
		}
	}

}
