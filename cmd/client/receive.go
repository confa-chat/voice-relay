package main

import (
	"context"
	"fmt"

	"github.com/royalcat/konfa/internal/codec"
	voicev1 "github.com/royalcat/konfa/internal/proto/gen/konfa/voice/v1"

	"github.com/gordonklaus/portaudio"
)

func receiveAudio(ctx context.Context, vclient voicev1.VoiceServiceClient, server, channel, user string) error {
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
	stream, err := portaudio.OpenStream(params, &buf)
	if err != nil {
		return err
	}
	stream.Start()
	defer stream.Close()

	dec, err := codec.NewDecoder(codec.CodecOpus)
	if err != nil {
		return fmt.Errorf("error creating opus stream: %w", err)
	}

	rec, err := vclient.Receive(ctx, &voicev1.ReceiveRequest{
		VoiceInfo: &voicev1.VoiceInfo{
			Codec:     voicev1.Codec_CODEC_OPUS,
			ServerId:  server,
			ChannelId: channel,
			UserId:    user,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating receive stream: %w", err)
	}

	println("start receiving audio for user ", user)

	for {
		msg, err := rec.Recv()
		if err != nil {
			return fmt.Errorf("error reading packet: %w", err)
		}

		vd := msg.GetVoiceData()
		if vd == nil {
			return fmt.Errorf("voice data not found in message")
		}

		data, err := dec.Decode(vd.Data)
		if err != nil {
			return fmt.Errorf("error reading from opus stream: %w", err)
		}

		n := copy(buf, data)
		buf = buf[:n]

		err = stream.Write()
		if err != nil {
			// return fmt.Errorf("error writing to portaudio stream: %w", err)
		}
	}
}
