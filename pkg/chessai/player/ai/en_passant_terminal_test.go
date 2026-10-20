package ai_test

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestQuiesceEnPassantOnlyMoveIsNotStalemate(t *testing.T) {
	parsed, err := analysis.ParseFEN("5n2/8/5k1K/6pP/8/8/8/8 w - g6 0 1")
	if err != nil {
		t.Fatal(err)
	}
	moves := parsed.Board.GetAllMovesUnShuffled(parsed.Active, parsed.Previous)
	if len(*moves) != 1 || analysis.MoveToUCI((*moves)[0]) != "h5g6" {
		t.Fatalf("expected sole legal move h5g6, got %v", *moves)
	}
	child := parsed.Board.Copy()
	last := board.MakeMove(&(*moves)[0], child)
	player := ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
	want := -player.Quiesce(child, ai.NegInf, ai.PosInf, parsed.Active^1, last)
	if want >= 0 {
		t.Fatalf("fixture must remain worse for White after en passant, got %d", want)
	}
	player = ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
	if got := player.Quiesce(parsed.Board, ai.NegInf, 0, parsed.Active, parsed.Previous); got >= 0 || got < want {
		t.Fatalf("en-passant-only position must not fail high on a false draw: got %d, capture score %d", got, want)
	}
	if player.Metrics.MovesConsidered == 0 {
		t.Fatal("quiescence must search the en passant capture")
	}
}

func TestEvaluationEnPassantOnlyMoveIsNotTerminal(t *testing.T) {
	for _, tt := range []struct {
		name, fen, move string
		inCheck         bool
	}{
		{"white-check", "8/8/8/6pP/5k1K/8/5n2/8 w - g6 0 1", "h5g6", true},
		{"white-stalemate", "5n2/8/5k1K/6pP/8/8/8/8 w - g6 0 1", "h5g6", false},
		{"black-check", "8/5N2/8/5K1k/6Pp/8/8/8 b - g3 0 1", "h4g3", true},
		{"black-stalemate", "8/8/8/8/6Pp/5K1k/8/5N2 b - g3 0 1", "h4g3", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := analysis.ParseFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			b := parsed.Board
			moves := b.GetAllMovesUnShuffled(parsed.Active, parsed.Previous)
			if len(*moves) != 1 || analysis.MoveToUCI((*moves)[0]) != tt.move || b.IsKingInCheck(parsed.Active) != tt.inCheck {
				t.Fatalf("expected sole legal escape %s, check=%v; got %v", tt.move, tt.inCheck, *moves)
			}
			player := ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
			terminal := ai.StalemateScore
			if tt.inCheck {
				terminal = ai.LossScore
			}
			// Alternate the same board's EP rights to cover terminal detection before caching.
			for i := 0; i < 2; i++ {
				if got := player.EvaluateBoard(b, parsed.Active, nil).TotalScore; got != terminal {
					t.Fatalf("without EP rights expected terminal score %d, got %d", terminal, got)
				}
				for _, got := range []int{
					player.EvaluateBoard(b, parsed.Active, parsed.Previous).TotalScore,
					ai.EvaluateBoardNoCache(b, parsed.Active, parsed.Previous).TotalScore,
				} {
					if got <= ai.LossScore || got >= 0 {
						t.Fatalf("EP escape must retain a nonterminal material disadvantage, got %d", got)
					}
				}
			}
			b.MovesSinceNoDraw = 100
			if got := player.EvaluateBoard(b, parsed.Active, parsed.Previous).TotalScore; got != ai.StalemateScore {
				t.Fatalf("EP escape must preserve the fifty-move draw instead of reporting mate, got %d", got)
			}
		})
	}
}
