package board_test

import (
	"reflect"
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
)

func TestIsCheckmatingMove(t *testing.T) {
	for _, tt := range []struct {
		name, fen, move string
		mate            bool
	}{
		{"white mate", "7k/5K2/6Q1/8/8/8/8/8 w - - 0 1", "g6g7", true},
		{"black mate", "8/8/8/8/8/6q1/5k2/7K b - - 0 1", "g3g2", true},
		{"discovered mate", "7k/5K2/8/8/8/8/7B/7R w - - 0 1", "h2g3", true},
		{"check with escape", "7k/8/6Q1/8/8/8/8/K7 w - - 0 1", "g6g7", false},
		{"stalemate", "7k/5K2/8/6Q1/8/8/8/8 w - - 0 1", "g5g6", false},
		{"castling mate", "4Nk2/8/2N3B1/3Q4/8/8/8/4K2R w K - 0 1", "e1g1", true},
		{"en passant check", "4k3/8/8/3pP3/8/8/8/K3R3 w - d6 0 1", "e5d6", false},
		{"queen promotion mate", "k7/2P5/1K6/8/8/8/8/8 w - - 0 1", "c7c8q", true},
		{"rook promotion mate", "k7/2P5/1K6/8/8/8/8/8 w - - 0 1", "c7c8r", true},
		{"bishop promotion", "k7/2P5/1K6/8/8/8/8/8 w - - 0 1", "c7c8b", false},
		{"knight promotion", "k7/2P5/1K6/8/8/8/8/8 w - - 0 1", "c7c8n", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := analysis.ParseFEN(tt.fen)
			if err != nil {
				t.Fatal(err)
			}
			selected, err := analysis.MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, tt.move)
			if err != nil {
				t.Fatal(err)
			}
			if got := parsed.Board.IsCheckmatingMove(selected); got != tt.mate {
				t.Fatalf("%s: got mate=%v, want %v", tt.move, got, tt.mate)
			}
			// Compare every legal alternative with full move application. This also
			// checks the fast rejection path without depending on search ordering.
			for _, move := range *parsed.Board.GetAllMoves(parsed.Active, parsed.Previous) {
				before := *parsed.Board
				full := parsed.Board.Copy()
				last := board.MakeMove(&move, full)
				want := full.IsKingInCheck(parsed.Active^1) && !full.HasLegalMove(parsed.Active^1, last)
				if got := parsed.Board.IsCheckmatingMove(move); got != want {
					t.Fatalf("%v: fast mate=%v, full mate=%v", move, got, want)
				}
				if !reflect.DeepEqual(parsed.Board, &before) {
					t.Fatalf("mate probe changed the original board for %v", move)
				}
			}
		})
	}
}
