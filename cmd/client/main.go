package main

import (
	"context"
	"flag"
	"fmt"
	"slices"

	voicev1 "github.com/confa-chat/voice-relay/internal/proto/gen/confa/voice/v1"
	"github.com/gordonklaus/portaudio"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	var host string
	flag.StringVar(&host, "host", "localhost:8081", "host")

	var server string
	flag.StringVar(&server, "server", "", "server")

	var channel string
	flag.StringVar(&channel, "channel", "", "channel")

	var user string
	flag.StringVar(&user, "user", "", "username")

	flag.Parse()

	if server == "" || channel == "" || user == "" {
		panic("user and other-user cannot be empty")
	}

	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(fmt.Errorf("error creating grpc client: %w", err))
	}

	ctx := context.Background()

	vclient := voicev1.NewVoiceRelayServiceClient(conn)

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		err := sendAudio(ctx, vclient, server, channel, user)
		if err != nil {
			println(fmt.Errorf("error sending audio: %w", err).Error())
			return err
		}
		return nil
	})

	group.Go(func() error {
		out, err := vclient.JoinChannel(ctx, &voicev1.JoinChannelRequest{
			ServerId:  server,
			ChannelId: channel,
		})
		if err != nil {
			return fmt.Errorf("error subscribing to channel state: %w", err)
		}

		userListened := []string{}

		for {
			state, err := out.Recv()
			if err != nil {
				return fmt.Errorf("error reading channel state: %w", err)
			}

			if state.GetUsersState() == nil {
				continue
			}

			for _, u := range state.GetUsersState().UserIds {
				if !slices.Contains(userListened, u) && u != user {
					go func() {
						err := receiveAudio(ctx, vclient, server, channel, u)
						if err != nil {
							panic(fmt.Errorf("error receiving audio: %w", err))
						}
						userListened = append(userListened, u)
					}()
				}
			}
		}
	})

	group.Wait()
}
