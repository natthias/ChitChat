package main

import (
	pb "ChitChatServer/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// --- config ---
	serverAddr := "localhost:6969"
	user := "alice"
	if len(os.Args) > 1 {
		user = os.Args[1]
	}

	// --- dial ---
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	c := pb.NewChitChatClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- join ---
	if _, err := c.JoinServer(ctx, &pb.JoinServerRequest{Username: user}); err != nil {
		log.Fatalf("join: %v", err)
	}
	log.Printf("joined as %s", user)

	// --- receive in background ---
	stream, err := c.ReceiveMessages(ctx, &pb.ReceiveMessagesRequest{Username: user})
	if err != nil {
		log.Fatalf("receive: %v", err)
	}
	go func() {
		for {
			m, err := stream.Recv()
			if err != nil {
				log.Printf("recv closed: %v", err)
				return
			}
			fmt.Printf("[%d] %s: %s\n", m.Timestamp, m.Sender, m.Message)
		}
	}()

	// --- handle Ctrl-C to leave ---
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		_, _ = c.LeaveServer(context.Background(), &pb.LeaveServerRequest{Username: user})
		os.Exit(0)
	}()

	// --- read stdin and publish ---
	sc := bufio.NewScanner(os.Stdin)
	fmt.Println("type messages and press Enter; Ctrl-C to quit")
	for sc.Scan() {
		text := sc.Text()
		if text == "" {
			continue
		}
		_, err := c.PublishMessage(ctx, &pb.ChatMessage{Sender: user, Message: text})
		if err != nil {
			log.Printf("publish: %v", err)
		}
	}
	// optional: leave on EOF (e.g., piped input)
	_, _ = c.LeaveServer(context.Background(), &pb.LeaveServerRequest{Username: user})
}
