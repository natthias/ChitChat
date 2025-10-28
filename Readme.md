# ChitChat
This program consists of a server and a client for the ChitChat messaging application

## Generating grpc
1. run `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/proto.proto` in the root of the project

## Running the program
1. Open a terminal and navigate to the server directory (e.g. `cd server`)
1. Start the server by running the appropriate command (e.g. `go run server.go`).
1. Open another terminal and navigate to the client directory (e.g. `cd client`)
1. Start the client by running the appropriate command (e.g., `go run client.go <username>`),
   Where `<username>` is the desired username for the client.
1. Once the client is connected, you can type messages in the client terminal.
1. The messages you type will be streamed to the server and broadcast to all other connected clients.
1. Other clients will see the messages in their respective terminals.
