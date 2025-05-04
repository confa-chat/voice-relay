package main

import (
	"context"
	"fmt"
	"io"

	"git.kmsign.ru/royalcat/konfa-voice/internal/codec"
	voicev1 "git.kmsign.ru/royalcat/konfa-voice/internal/proto/gen/konfa/voice/v1"

	"github.com/gordonklaus/portaudio"
)

func sendAudio(ctx context.Context, vclient voicev1.VoiceServiceClient, server, channel, user string) error {
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

	enc, err := codec.NewEncoder(codec.CodecOpus)
	if err != nil {
		return fmt.Errorf("error creating opus encoder: %w", err)
	}

	w, err := vclient.SpeakToChannel(ctx)
	if err != nil {
		return fmt.Errorf("error creating send stream: %w", err)
	}

	err = w.Send(&voicev1.SpeakToChannelRequest{
		Request: &voicev1.SpeakToChannelRequest_VoiceInfo{
			VoiceInfo: &voicev1.VoiceInfo{
				Codec:     voicev1.AudioCodec_AUDIO_CODEC_OPUS,
				ServerId:  server,
				ChannelId: channel,
				UserId:    user,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error sending first message: %w", err)
	}

	println("start sending audio")

	for {
		err := stream.Read()
		if err != nil {
			return fmt.Errorf("error reading audio: %w", err)
		}

		data, err := enc.Encode(buf)
		if err != nil {
			return fmt.Errorf("error encoding audio: %w", err)
		}

		err = w.Send(&voicev1.SpeakToChannelRequest{
			Request: &voicev1.SpeakToChannelRequest_VoiceData{
				VoiceData: &voicev1.VoiceData{
					Data: data,
				},
			},
		})
		if err != nil && err != io.EOF {
			return fmt.Errorf("error sending audio: %w", err)
		}
	}

}
