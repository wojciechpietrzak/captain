package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb "captain/src/pairing_engine"
	"captain/src/pairing_engine/server"

	"google.golang.org/protobuf/proto"
)

func convertTablesToGames(tables []*pb.Table) []*pb.Game {
	var games []*pb.Game
	for _, table := range tables {
		games = append(games, &pb.Game{
			Table:       table,
			WhiteResult: nil, // Default: game in progress
			BlackResult: nil, // Default: game in progress
		})
	}
	return games
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run wrapper.go <tournament.json>")
	}

	tournamentFile := os.Args[1]

	// Load tournament data from JSON
	data, err := ioutil.ReadFile(tournamentFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var tournament pb.Tournament
	if err := proto.Unmarshal(data, &tournament); err != nil {
		log.Fatalf("Failed to unmarshal JSON to protobuf: %v", err)
	}

	// Use the server implementation to calculate pairing
	engine := server.PairingEngineServer{}
	resp, err := engine.CalculatePairing(nil, &pb.CalculatePairingRequest{Tournament: &tournament})
	if err != nil {
		log.Fatalf("Failed to calculate pairing: %v", err)
	}

	// Update tournament with the new pairing
	newRound := pb.Round{
		Games: convertTablesToGames(resp.Pairing.Tables),
		Byes:  resp.Pairing.EmptyTables,
	}
	tournament.Rounds = append(tournament.Rounds, &newRound)

	// Save updated tournament to JSON
	updatedData, err := proto.Marshal(&tournament)
	if err != nil {
		log.Fatalf("Failed to marshal protobuf to JSON: %v", err)
	}
	if err := ioutil.WriteFile(tournamentFile, updatedData, 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Println("Tournament updated successfully.")
}