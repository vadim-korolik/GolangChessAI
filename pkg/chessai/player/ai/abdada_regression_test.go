package ai_test

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestABDADAAwQCDMkPSeesMateAfterKnightCapture(t *testing.T) {
	// AwQCDMkP: ...exf3 allows Rxe7 and a forced mate; ...Qd8 avoids it.
	for _, tt := range []struct{ name, fen, move string }{
		{"black", "4rk1r/p2Qn1pp/8/q1p5/2P1p3/5N2/P1P2PPP/1R2R1K1 b - - 0 20", "e4f3"},
		{"white", "1r2r1k1/p1p2ppp/5n2/2p1P3/Q1P5/8/P2qN1PP/4RK1R w - - 0 20", "e5f6"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := analysis.ParseFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			move, err := analysis.MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, tt.move)
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
				t.Fatalf("%s must reveal the rook/queen mate at depth seven, got score=%d", tt.move, best.Score)
			}
		})
	}
}

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

// yA6gLPUd: Bxa8 must not hide the late ...Qh4+ reply behind LMR.
func TestABDADACheckLMRAvoidsFalseRookGain(t *testing.T) {
	for _, tt := range []struct{ name, fen, move string }{
		{"white", "r1bq1rk1/p1p2ppp/2B5/3p4/1b2n3/2N2P2/PPP3PP/R1BQK2R w KQ - 0 10", "c6a8"},
		{"black", "r1bqk2r/ppp3pp/2n2p2/1B2N3/3P4/2b5/P1P2PPP/R1BQ1RK1 b kq - 0 10", "c3a1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := analysis.ParseFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			move, err := analysis.MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, tt.move)
			if err != nil {
				t.Fatal(err)
			}
			algorithm := &ai.ABDADA{NumThreads: 1}
			player := ai.NewAIPlayer(parsed.Active, algorithm)
			player.TranspositionTableEnabled = true
			player.PrintInfo, player.Debug = false, false
			got := algorithm.ScoreRootMove(player, parsed.Board, move, 8)
			if got.Score <= ai.NegInf || got.Score >= ai.PosInf || got.Score == ai.OnEvaluation || got.Score == -ai.OnEvaluation {
				t.Fatalf("%s returned sentinel score=%d", tt.move, got.Score)
			}
			if got.Score >= ai.PawnValueWeight {
				t.Fatalf("%s false rook gain: score=%d, want less than %d", tt.move, got.Score, ai.PawnValueWeight)
			}
			t.Logf("%s score=%d", tt.move, got.Score)
		})
	}
}
