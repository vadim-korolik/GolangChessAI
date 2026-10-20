package ai_test

import (
	"fmt"
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestQuiesceQuietEvaluationPreservesDrawCountersAndCache(t *testing.T) {
	for _, side := range []string{"w", "b"} {
		parsed, err := analysis.ParseFEN(fmt.Sprintf("6kr/5p2/8/8/8/8/2P5/R5K1 %s - - 0 20", side))
		if err != nil {
			t.Fatal(err)
		}
		player := ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
		for _, halfmoves := range []int{0, 12, 99, 100, 12} {
			parsed.Board.MovesSinceNoDraw = halfmoves
			want := player.EvaluateBoard(parsed.Board, parsed.Active, parsed.Previous).TotalScore
			if halfmoves == 100 && want != ai.StalemateScore {
				t.Fatalf("fifty-move boundary must draw, got %d", want)
			}
			got := player.Quiesce(parsed.Board, ai.NegInf, ai.PosInf, parsed.Active, parsed.Previous)
			if got != want {
				t.Fatalf("side=%s halfmoves=%d: quiet quiescence=%d, full evaluation=%d", side, halfmoves, got, want)
			}
		}
		parsed.Board.CurrentPositionRepeats = 2
		if got := player.Quiesce(parsed.Board, ai.NegInf, ai.PosInf, parsed.Active, parsed.Previous); got != ai.StalemateScore {
			t.Fatalf("repetition must draw instead of reusing a cached static score, got %d", got)
		}
	}
}

func TestQuiesceTerminalPrecedenceAtDrawBoundary(t *testing.T) {
	for _, tt := range []struct {
		name, fen string
		want      int
	}{
		{"mate", "7k/6Q1/5K2/8/8/8/8/8 b - - 100 50", ai.LossScore},
		{"stalemate", "7k/5K2/6Q1/8/8/8/8/8 b - - 100 50", ai.StalemateScore},
		{"insufficient", "4k3/8/8/8/8/8/8/4K3 w - - 0 1", ai.StalemateScore},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := analysis.ParseFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			parsed.Board.CurrentPositionRepeats = 2
			player := ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
			if got := player.Quiesce(parsed.Board, ai.NegInf, ai.PosInf, parsed.Active, parsed.Previous); got != tt.want {
				t.Fatalf("terminal priority: got %d, want %d", got, tt.want)
			}
		})
	}
}
