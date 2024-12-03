package voice

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/royalcat/konfa/internal/codec"
	voicev1 "github.com/royalcat/konfa/internal/proto/gen/konfa/voice/v1"
	"google.golang.org/grpc"
)

func NewService() *Service {
	return &Service{
		servers: make(map[string]*Server),
	}
}

type Service struct {
	servers map[string]*Server
	// voicev1.UnimplementedVoiceServiceServer
}

var _ voicev1.VoiceServiceServer = (*Service)(nil)

func (s *Service) server(name string) *Server {
	if _, ok := s.servers[name]; !ok {
		s.servers[name] = NewServer()
	}

	return s.servers[name]
}

// SubscribeChannelState implements voicev1.VoiceServiceServer.
func (s *Service) SubscribeChannelState(req *voicev1.SubscribeChannelStateRequest, out grpc.ServerStreamingServer[voicev1.SubscribeChannelStateResponse]) error {
	vc := s.server(req.ServerId).channel(req.ChannelId)

	err := out.Send(&voicev1.SubscribeChannelStateResponse{
		Users: vc.Users(),
	})
	if err != nil {
		return err
	}

	oldUsers := vc.Users()
	slices.Sort(oldUsers)

	for range time.NewTicker(time.Millisecond).C {
		newUsers := vc.Users()
		slices.Sort(newUsers)

		if !slices.Equal(oldUsers, newUsers) {
			err := out.Send(&voicev1.SubscribeChannelStateResponse{
				Users: vc.Users(),
			})
			if err != nil {
				return err
			}
			oldUsers = newUsers
		}
	}

	return nil
}

// Send implements voicev1.VoiceServiceServer.
func (s *Service) Send(in grpc.ClientStreamingServer[voicev1.SendRequest, voicev1.SendResponse]) error {
	msg, err := in.Recv()
	if err != nil {
		return fmt.Errorf("error receiving first message: %w", err)
	}

	vi := msg.GetVoiceInfo()
	if vi == nil {
		return fmt.Errorf("voice info not found in first message")
	}

	vc := s.server(vi.ServerId).channel(vi.ChannelId)

	cdc, err := mapCodec(vi.Codec)
	if err != nil {
		return fmt.Errorf("error initializing codec: %w", err)
	}

	dec, err := codec.NewDecoder(cdc)
	if err != nil {
		return fmt.Errorf("error creating encoder: %w", err)
	}

	for {
		msg, err := in.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("error receiving  message: %w", err)
		}

		data := msg.GetVoiceData()
		if data == nil {
			return fmt.Errorf("voice data not found in message")
		}

		sound, err := dec.Decode(data.Data)
		if err != nil {
			return fmt.Errorf("error decoding voice data: %w", err)
		}

		vc.Send(vi.UserId, sound)
	}

}

// Receive implements voicev1.VoiceServiceServer.
func (s *Service) Receive(req *voicev1.ReceiveRequest, out grpc.ServerStreamingServer[voicev1.ReceiveResponse]) error {
	vi := req.VoiceInfo
	if vi == nil {
		return fmt.Errorf("voice info not found in first message")
	}

	vc := s.server(vi.ServerId).channel(vi.ChannelId)

	cdc, err := mapCodec(vi.Codec)
	if err != nil {
		return fmt.Errorf("error initializing codec: %w", err)
	}

	enc, err := codec.NewEncoder(cdc)
	if err != nil {
		return fmt.Errorf("error creating encoder: %w", err)
	}

	for v := range vc.Listener(vi.UserId).Out() {
		data, err := enc.Encode(v)
		if err != nil {
			return fmt.Errorf("error encoding voice data: %w", err)
		}

		err = out.Send(&voicev1.ReceiveResponse{
			Response: &voicev1.ReceiveResponse_VoiceData{
				VoiceData: &voicev1.VoiceData{
					Data: data,
				},
			},
		})
		if err != nil && err != io.EOF {
			return fmt.Errorf("error sending message: %v", err)
		}
	}

	return nil
}
