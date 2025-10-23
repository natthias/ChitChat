package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type ChitChat_server struct {
	proto.UnimplementedChitChatClientsServer
	clients []string
}

func (s *ChitChat_server) GetClients(ctx context.Context, in *proto.Empty) (*proto.Clients, error) {
	return &proto.Clients{Clients: s.clients}, nil
}

func main() {
	server := ChitChat_server{clients: []string{}}
	log.Print("ChitChat server has been initiated 🚀")

	server.clients = append(server.clients, "John")

	server.start_server()
	log.Print("ChitChat server has terminated 🪿")
}

func (s *ChitChat_server) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChitChatClientsServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}

}
