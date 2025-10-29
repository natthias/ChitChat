package main

import (
	proto "ChitChatServer/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	mu      sync.Mutex
	clients map[string]chan *proto.ChatMessage
	clock   uint64
}

func main() {
	server := ChitChatServer{clients: make(map[string]chan *proto.ChatMessage)}
	log.Print("ChitChat server has been initiated 🚀")

	server.start_server()
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChitChatServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}

}

// JoinServer handles a new client joining the chat.
func (s *ChitChatServer) JoinServer(ctx context.Context, req *proto.JoinServerRequest) (*proto.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If the client is already connected, do nothing.
	if _, exists := s.clients[req.Username]; exists {
		return &proto.Empty{}, nil
	}

	// Create a buffered channel for the new client.
	ch := make(chan *proto.ChatMessage, 32)
	s.clients[req.Username] = ch

	// Increment the Lamport clock.
	s.clock = max(s.clock, req.Timestamp) + 1

	// Broadcast a “user joined” message to all clients.
	joinMsg := &proto.ChatMessage{
		Sender:    "Server",
		Message:   fmt.Sprintf("%s joined Chat", req.Username),
		Timestamp: s.clock,
	}
	for _, clientCh := range s.clients {
		clientCh <- joinMsg
	}
	log.Printf("[Server] join: user=%s lt=%d", req.Username, s.clock)

	return &proto.Empty{}, nil
}

// LeaveServer handles a client leaving the chat.
func (s *ChitChatServer) LeaveServer(ctx context.Context, req *proto.LeaveServerRequest) (*proto.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove the client’s channel.
	ch, exists := s.clients[req.Username]
	if !exists {
		return &proto.Empty{}, nil
	}
	delete(s.clients, req.Username)
	close(ch)

	// Increment the Lamport clock.
	s.clock = max(s.clock, req.Timestamp) + 1

	// Broadcast a “user left” message to the remaining clients.
	leaveMsg := &proto.ChatMessage{
		Sender:    "Server",
		Message:   fmt.Sprintf("%s left Chat", req.Username),
		Timestamp: s.clock,
	}
	for _, clientCh := range s.clients {
		clientCh <- leaveMsg
	}
	log.Printf("[Server] leave: user=%s lt=%d", req.Username, s.clock)

	return &proto.Empty{}, nil
}

// PublishMessage broadcasts a chat message to all clients.
func (s *ChitChatServer) PublishMessage(ctx context.Context, msg *proto.ChatMessage) (*proto.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Increment the Lamport clock and stamp the message.
	s.clock = max(s.clock, msg.Timestamp) + 1
	msg.Timestamp = s.clock

	// Fan the message out to all client channels.
	for _, ch := range s.clients {

		select {
		case ch <- msg:
		default:
			// Optional: drop or log if a client's buffer is full.
		}
	}
	log.Printf("[Server] publish (%s): from=%s lt=%d", msg.Sender, msg.Message, s.clock)

	return &proto.Empty{}, nil
}

// ReceiveMessages streams messages to a specific client.
func (s *ChitChatServer) ReceiveMessages(req *proto.ReceiveMessagesRequest, stream proto.ChitChat_ReceiveMessagesServer) error {
	s.mu.Lock()
	ch, ok := s.clients[req.Username]
	s.mu.Unlock()
	if !ok {
		// client hasn't joined or was removed; return silently
		return nil
	}

	// Read from the client's channel and stream messages back.
	for msg := range ch {
		s.mu.Lock()
		s.clock++
		msg.Timestamp = s.clock
		if err := stream.Send(msg); err != nil {
			return err
		}
		log.Printf("[Server] delivered: to=%s lt=%d", req.Username, msg.Timestamp)
		s.mu.Unlock()
	}
	return nil
}
