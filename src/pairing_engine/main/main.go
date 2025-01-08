package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "captain/src/pairing_engine"
	server "captain/src/pairing_engine/server"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPairingEngineServer(grpcServer, &server.PairingEngineServer{})

	log.Println("Server is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}