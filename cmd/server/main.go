package main

import (
	"fmt"
	"log"
	"net"

	voicev1 "git.kmsign.ru/royalcat/konfa-voice/internal/proto/gen/konfa/voice/v1"
	"git.kmsign.ru/royalcat/konfa-voice/internal/voice"
	"google.golang.org/grpc"
)

func main() {
	voiceService := voice.NewService()
	port := 8081

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	voicev1.RegisterVoiceServiceServer(grpcServer, voiceService)

	println("Server is running on port", port)
	panic(grpcServer.Serve(lis))
}
