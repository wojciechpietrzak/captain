package server

import (
	pb "captain/src/pairing_engine" // Import the generated Protobuf package
	"context"
	"fmt"
	"log"
	"sort"
)

type PairingEngineServer struct {
	pb.UnimplementedPairingEngineServer
}

// Implement the CalculatePairing RPC
func (s *PairingEngineServer) CalculatePairing(ctx context.Context, req *pb.CalculatePairingRequest) (*pb.CalculatePairingResponse, error) {
	tables := []*pb.Table{}
	emptyTables := []*pb.EmptyTable{}

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

	var currentRoundPairedPlayers []*pb.Player

	for _, player := range req.Tournament.Players {
		withdrawal := getWithdrawal(player, currentRound)
		if withdrawal == nil {
			currentRoundPairedPlayers = append(currentRoundPairedPlayers, player)
		} else {
			emptyTables = append(emptyTables, &pb.EmptyTable{
				TableNo:       0, // to be adjusted later
				PlayerStartNo: player.StartNo,
				Bye:           withdrawal,
			})
		}
	}

	log.Printf("Current round paired players: %v", currentRoundPairedPlayers)

	var brackets = initBrackets(req.Tournament)

	log.Printf("Initial brackets for round %v pairing: %v", currentRound, brackets)

	// Last step: correct table numbers

	for i, table := range tables {
		table.TableNo = int32(i + 1)
	}
	for i, emptyTable := range emptyTables {
		emptyTable.TableNo = int32(i + 1 + len(tables))
	}

	return &pb.CalculatePairingResponse{
		Pairing: &pb.Pairing{
			Tables:      tables, // TODO: implement me
			EmptyTables: emptyTables,
		},
	}, nil
}

func getWithdrawal(player *pb.Player, round int) *pb.Bye {
	for _, withdrawal := range player.Withdrawals {
		if int(withdrawal.RoundNo) == round {
			return withdrawal.Bye
		}
	}
	return nil
}

type ScoreGroup struct {
	Score   float64
	Players []*pb.Player
}

type Bracket = []ScoreGroup

func initBrackets(tournament *pb.Tournament) []Bracket {
	// the initial brackets are always homogenous,
	// i.e. they contain only one scoregroup each

	playerScores := make(map[int32]float64)

	for _, player := range tournament.Players {
		playerScores[player.StartNo] = 0
	}

	for _, round := range tournament.Rounds {
		// Set to ensure a player does not appear twice in a round
		playersInRound := make(map[int32]bool)

		// Process games
		for _, game := range round.Games {
			if game.Table.WhitePlayerStartNo != 0 {
				startNo := game.Table.WhitePlayerStartNo
				if playersInRound[startNo] {
					log.Panicf("Player %d appears multiple times in a single round", startNo)
				}
				playersInRound[startNo] = true

				points := float64(0.5) // Default value if result is missing
				if game.WhiteResult.Points == 0 || game.WhiteResult.Points == 0.5 || game.WhiteResult.Points == 1 {
					points = float64(game.WhiteResult.Points)
				} else {
					log.Panicf("Invalid result points for player %d: %f", startNo, game.WhiteResult.Points)
				}
				playerScores[startNo] += points
			}

			if game.Table.BlackPlayerStartNo != 0 {
				startNo := game.Table.BlackPlayerStartNo
				if playersInRound[startNo] {
					log.Panicf("Player %d appears multiple times in a single round", startNo)
				}
				playersInRound[startNo] = true

				points := float64(0.5) // Default value if result is missing
				if game.BlackResult.Points == 0 || game.BlackResult.Points == 0.5 || game.BlackResult.Points == 1 {
					points = float64(game.BlackResult.Points)
				} else {
					log.Panicf("Invalid result points for player %d: %f", startNo, game.BlackResult.Points)
				}
				playerScores[startNo] += points
			}
		}

		// Process byes
		for _, bye := range round.Byes {
			startNo := bye.PlayerStartNo
			if playersInRound[startNo] {
				log.Panicf("Player %d appears multiple times in a single round", startNo)
			}
			playersInRound[startNo] = true

			playerScores[startNo] += float64(bye.Bye.ByeVal)
		}
	}

	scoreGroups := make(map[float64][]*pb.Player)
	for _, player := range tournament.Players {
		score := playerScores[player.StartNo]
		scoreGroups[score] = append(scoreGroups[score], player)
	}

	var brackets []Bracket
	var sortedScores []float64
	for score := range scoreGroups {
		sortedScores = append(sortedScores, score)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(sortedScores)))

	for _, score := range sortedScores {
		players := scoreGroups[score]
		sort.Slice(players, func(i, j int) bool {
			return players[i].StartNo < players[j].StartNo
		})
		brackets = append(brackets, []ScoreGroup{
			{
				Score:   score,
				Players: players,
			},
		})
	}

	return brackets
}
