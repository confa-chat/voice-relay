package main

import (
	"fmt"
	"io"
	"konfa/voip/internal/codec"

	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
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

	enc, err := opus.NewEncoder(codec.SampleRate, codec.Channels, opus.AppVoIP)
	if err != nil {
		return fmt.Errorf("error creating opus encoder: %w", err)
	}

	println("start sending audio")

	for {
		err := stream.Read()
		if err != nil {
			return err
		}

		out := make([]byte, codec.MaxPacketSize)
		n, err := enc.EncodeFloat32(buf, out)
		if err != nil {
			return fmt.Errorf("error encoding audio: %w", err)
		}

		codec.WritePacket(w, out[:n])
	}

}
