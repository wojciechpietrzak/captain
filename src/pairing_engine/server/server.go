package server

import (
	pb "captain/src/pairing_engine" // Import the generated Protobuf package
	"context"
	"fmt"
	"log"
)

type PairingEngineServer struct {
	pb.UnimplementedPairingEngineServer
}

// Implement the CalculatePairing RPC
func (s *PairingEngineServer) CalculatePairing(ctx context.Context, req *pb.CalculatePairingRequest) (*pb.CalculatePairingResponse, error) {
	allRounds := int(req.Tournament.AllRoundsNo)
	currentRound := len(req.Tournament.Rounds) + 1

	if currentRound > allRounds {
		log.Println("All rounds already paired")
		return nil, fmt.Errorf("all rounds already paired")
	}

	applyLast := currentRound == allRounds

	log.Printf("Pairing round %v of %v.", currentRound, allRounds)
	if applyLast {
		log.Println("Applying last round special rules.")
	}

	var currentRoundPlayers []*pb.Player
	var withdrawn bool

	for _, player := range req.Tournament.Players {
		withdrawn = false
		for _, withdrawal := range player.Withdrawals {
			if (int)(withdrawal.RoundNo) == currentRound {
				withdrawn = true
				break
			}
		}
		if !withdrawn {
			currentRoundPlayers = append(currentRoundPlayers, player)
		}
	}

	log.Printf("Current round players: %v", currentRoundPlayers)

	return &pb.CalculatePairingResponse{
		Pairing: &pb.Pairing{
			Tables:      []*pb.Table{},      // TODO: implement me
			EmptyTables: []*pb.EmptyTable{}, // TODO: implement me
		},
	}, nil
}
