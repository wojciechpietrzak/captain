package server

import (
	pb "captain/src/pairing_engine" // Import the generated Protobuf package
	"context"
)

type PairingEngineServer struct {
	pb.UnimplementedPairingEngineServer
}

// Implement the CalculatePairing RPC
func (s *PairingEngineServer) CalculatePairing(ctx context.Context, req *pb.CalculatePairingRequest) (*pb.CalculatePairingResponse, error) {
	// Logic for pairing goes here
	return &pb.CalculatePairingResponse{
		Pairing: &pb.Pairing{
			Tables:      []*pb.Table{},      // TODO: implement me
			EmptyTables: []*pb.EmptyTable{}, // TODO: implement me
		},
	}, nil
}
