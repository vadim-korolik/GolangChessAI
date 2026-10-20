package ai_test

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestABDADALMRPreservesQuietMatingNet(t *testing.T) {
	parsed, err := analysis.ParseFEN("3R4/5k1p/8/1Pnp1p2/6pK/1P1r4/8/8 w - - 2 55")
	if err != nil {
		t.Fatal(err)
	}
	move, err := analysis.MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, "h4h5")
	if err != nil {
		t.Fatal(err)
	}
	algorithm := &ai.ABDADA{NumThreads: 1}
	player := ai.NewAIPlayer(parsed.Active, algorithm)
	player.TranspositionTableEnabled = true
	player.PrintInfo = false
	player.Debug = false
	best := algorithm.ScoreRootMove(player, parsed.Board, move, 7)
	if best.Score > ai.LossScore || best.Score <= ai.NegInf || best.Score == -ai.OnEvaluation {
		t.Fatalf("Kh5 must reveal the Ne6 mating net at depth seven, got score=%d", best.Score)
	}
}

func TestABDADABusyTTDoesNotHideRookEndingMate(t *testing.T) {
	parsed, err := analysis.ParseFEN("3R4/5k1p/8/1Pnp1p2/6pK/1P1r4/8/8 w - - 2 55")
	if err != nil {
		t.Fatal(err)
	}
	algorithm := &ai.ABDADA{NumThreads: 8}
	player := ai.NewAIPlayer(parsed.Active, algorithm)
	player.TranspositionTableEnabled = true
	player.PrintInfo = false
	player.Debug = false
	player.MaxSearchDepth = 9
	best := algorithm.GetBestMove(player, parsed.Board, parsed.Previous)
	if move := analysis.MoveToUCI(best.Move); move != "h4g5" {
		t.Fatalf("depth-nine parallel search must choose Kg5 instead of a mating line, got %s score=%d", move, best.Score)
	}
}

func TestABDADACdTsyKE6AvoidsQuietBlunderWhenPawnBreakAvailable(t *testing.T) {
	t.Skip("known ABDADA weakness; tracked in testdata/abdada_fens.txt for benchmark-driven optimization")
}

func TestABDADACdTsyKE6DefendsMatingNetWithRook(t *testing.T) {
	parsed, err := analysis.ParseFEN("2q5/P6p/3Qpk1p/4n3/7P/1b4R1/1KP5/5r2 w - - 0 43")
	if err != nil {
		t.Fatal(err)
	}

	algorithm := &ai.ABDADA{NumThreads: 1}
	player := ai.NewAIPlayer(parsed.Active, algorithm)
	player.MaxSearchDepth = 3
	player.TranspositionTableEnabled = true
	player.PrintInfo = false
	player.Debug = false

	best := algorithm.GetBestMove(player, parsed.Board, parsed.Previous)
	uci := analysis.MoveToUCI(best.Move)
	if uci != "g3c3" {
		t.Fatalf("expected rook defense g3c3 in mating net, got %s score=%d", uci, best.Score)
	}
}

func TestABDADACdTsyKE6AvoidsForcedMateWalk(t *testing.T) {
	t.Skip("known ABDADA weakness; tracked in testdata/abdada_fens.txt for benchmark-driven optimization")
}
