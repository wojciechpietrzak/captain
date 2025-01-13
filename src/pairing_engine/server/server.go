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

	history := getCards(req.Tournament)

	log.Printf("History: %+v", history)

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

				points := 0.5 // Default value if result is missing
				if game.WhiteResult.Points == 0 || game.WhiteResult.Points == 0.5 || game.WhiteResult.Points == 1 {
					points = game.WhiteResult.Points
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

				points := 0.5 // Default value if result is missing
				if game.BlackResult.Points == 0 || game.BlackResult.Points == 0.5 || game.BlackResult.Points == 1 {
					points = game.BlackResult.Points
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

			playerScores[startNo] += bye.Bye.ByeVal
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

const (
	NoOpponent = 0

	NoColour    = 0
	WhiteColour = 1
	BlackColour = -1

	NoFloat   = 0
	UpFloat   = 1
	DownFloat = -1
)

type PlayerRoundCard struct {
	OpponentNo int32
	RoundNo    int
	Colour     int
	Result     float64
	Float      int
	Progress   float64
}

type PlayerCard struct {
	Player           *pb.Player
	History          []PlayerRoundCard
	ColourDifference int
	LastColour       int
}

func getCards(tournament *pb.Tournament) map[int32]PlayerCard {
	// Initialize player histories
	playersCards := make(map[int32]PlayerCard)
	for _, player := range tournament.Players {
		playersCards[player.StartNo] = PlayerCard{
			Player:           player,
			History:          []PlayerRoundCard{},
			ColourDifference: 0,
			LastColour:       0,
		}
	}

	// Iterate through rounds to fill player histories
	for roundIdx, round := range tournament.Rounds {
		roundNo := roundIdx + 1

		// Map to track players already paired in this round
		pairedPlayers := make(map[int32]bool)

		// Process games
		for _, game := range round.Games {
			white := game.Table.WhitePlayerStartNo
			black := game.Table.BlackPlayerStartNo

			// Add to white player's history
			if whiteHistory, ok := playersCards[white]; ok {
				points := 0.5
				if game.WhiteResult != nil {
					points = game.WhiteResult.Points
				}
				progress := points
				if roundIdx > 0 {
					progress += whiteHistory.History[roundIdx-1].Progress
				}
				floatVal := NoFloat
				if roundIdx > 0 {
					if blackHistory, ok := playersCards[black]; ok {
						prevWhiteProgress := whiteHistory.History[roundIdx-1].Progress
						prevBlackProgress := blackHistory.History[roundIdx-1].Progress
						if prevWhiteProgress > prevBlackProgress {
							floatVal = DownFloat
						} else if prevWhiteProgress < prevBlackProgress {
							floatVal = UpFloat
						}
					}
				}
				whiteHistory.History = append(whiteHistory.History, PlayerRoundCard{
					OpponentNo: black,
					RoundNo:    roundNo,
					Colour:     WhiteColour,
					Result:     points,
					Float:      floatVal,
					Progress:   progress,
				})
				whiteHistory.ColourDifference += WhiteColour
				whiteHistory.LastColour = WhiteColour
				playersCards[white] = whiteHistory
				pairedPlayers[white] = true
			}

			// Add to black player's history
			if blackHistory, ok := playersCards[black]; ok {
				points := 0.5
				if game.BlackResult != nil {
					points = game.BlackResult.Points
				}
				progress := points
				if roundIdx > 0 {
					progress += blackHistory.History[roundIdx-1].Progress
				}
				floatVal := NoFloat
				if roundIdx > 0 {
					if whiteHistory, ok := playersCards[white]; ok {
						prevBlackProgress := blackHistory.History[roundIdx-1].Progress
						prevWhiteProgress := whiteHistory.History[roundIdx-1].Progress
						if prevBlackProgress > prevWhiteProgress {
							floatVal = DownFloat
						} else if prevBlackProgress < prevWhiteProgress {
							floatVal = UpFloat
						}
					}
				}
				blackHistory.History = append(blackHistory.History, PlayerRoundCard{
					OpponentNo: white,
					RoundNo:    roundNo,
					Colour:     BlackColour,
					Result:     points,
					Float:      floatVal,
					Progress:   progress,
				})
				blackHistory.ColourDifference += BlackColour
				blackHistory.LastColour = BlackColour
				playersCards[black] = blackHistory
				pairedPlayers[black] = true
			}
		}

		// Process byes
		for _, bye := range round.Byes {
			playerNo := bye.PlayerStartNo
			if playerHistory, ok := playersCards[playerNo]; ok {
				points := bye.Bye.ByeVal
				floatVal := NoFloat
				if points > 0 {
					floatVal = DownFloat
				}
				progress := points
				if roundIdx > 0 {
					progress += playerHistory.History[roundIdx-1].Progress
				}
				playerHistory.History = append(playerHistory.History, PlayerRoundCard{
					OpponentNo: NoOpponent,
					RoundNo:    roundNo,
					Colour:     NoColour,
					Result:     points,
					Float:      floatVal,
					Progress:   progress,
				})
				playersCards[playerNo] = playerHistory
				pairedPlayers[playerNo] = true
			}
		}

		// Add missing players with default values
		for _, player := range tournament.Players {
			if !pairedPlayers[player.StartNo] {
				playerHistory := playersCards[player.StartNo]
				progress := 0.0
				if roundIdx > 0 {
					progress = playerHistory.History[roundIdx-1].Progress
				}
				playerHistory.History = append(playerHistory.History, PlayerRoundCard{
					OpponentNo: NoOpponent,
					RoundNo:    roundNo,
					Colour:     NoColour,
					Result:     0,
					Float:      NoFloat,
					Progress:   progress,
				})
				playersCards[player.StartNo] = playerHistory
			}
		}
	}

	return playersCards
}
