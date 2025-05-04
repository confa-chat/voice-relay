package voice

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"git.kmsign.ru/royalcat/konfa-voice/internal/codec"
	voicev1 "git.kmsign.ru/royalcat/konfa-voice/internal/proto/gen/konfa/voice/v1"
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

func (s *Service) JoinChannel(req *voicev1.JoinChannelRequest, out grpc.ServerStreamingServer[voicev1.JoinChannelResponse]) error {
	vc := s.server(req.ServerId).channel(req.ChannelId)

	vc.AddUser(req.UserId)
	defer vc.RemoveUser(req.UserId)

	err := out.Send(&voicev1.JoinChannelResponse{
		State: &voicev1.JoinChannelResponse_UsersState{
			UsersState: &voicev1.UsersState{
				UserIds: vc.Users(),
			},
		},
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
			err := out.Send(&voicev1.JoinChannelResponse{
				State: &voicev1.JoinChannelResponse_UsersState{
					UsersState: &voicev1.UsersState{
						UserIds: vc.Users(),
					},
				},
			})
			if err != nil {
				return err
			}
			oldUsers = newUsers
		}
	}

	return nil
}

func (s *Service) SpeakToChannel(in grpc.ClientStreamingServer[voicev1.SpeakToChannelRequest, voicev1.SpeakToChannelResponse]) error {
	msg, err := in.Recv()
	if err != nil {
		return fmt.Errorf("error receiving first message: %w", err)
	}

	vi := msg.GetVoiceInfo()
	if vi == nil {
		return fmt.Errorf("voice info not found in first message")
	}

	vc := s.server(vi.ServerId).channel(vi.ChannelId)

	if !slices.Contains(vc.Users(), vi.UserId) {
		return fmt.Errorf("user not in channel")
	}

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
func (s *Service) ListenToUser(req *voicev1.ListenToUserRequest, out grpc.ServerStreamingServer[voicev1.ListenToUserResponse]) error {
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

		err = out.Send(&voicev1.ListenToUserResponse{
			Response: &voicev1.ListenToUserResponse_VoiceData{
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
