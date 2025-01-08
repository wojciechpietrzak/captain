package main

import (
	"log"
	"net"

	pb "captain/src/pairing_engine" // Import the generated Protobuf package
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type PairingEngineServer struct {
	pb.UnimplementedPairingEngineServer
}

// Implement the CalculatePairing RPC
func (s *PairingEngineServer) CalculatePairing(ctx context.Context, req *pb.CalculatePairingRequest) (*pb.CalculatePairingResponse, error) {
	// Logic for pairing goes here
	return &pb.CalculatePairingResponse{}, nil
}

func main() {
	// Listen on port 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterPairingEngineServer(grpcServer, &PairingEngineServer{})

	// Register reflection service on gRPC server (optional, for debugging)
	reflection.Register(grpcServer)

	log.Println("Server is listening on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
