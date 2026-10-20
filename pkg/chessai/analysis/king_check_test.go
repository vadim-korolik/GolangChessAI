package analysis

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
)

func assertKingCheckMatchesAttackMap(t *testing.T, b *board.Board) {
	t.Helper()
	for _, side := range []color.Color{color.White, color.Black} {
		attacks := b.GetAllAttackableMoves(side ^ 1)
		want := attacks.IsLocationSet(b.KingLocations[side])
		if got := b.IsKingInCheck(side); got != want {
			t.Fatalf("check for side %d = %t, want %t; position %s", side, got, want, BoardToFEN(b, side, nil, 1))
		}
	}
}

func TestKingCheckMatchesAttackMapOnStrengthCorpus(t *testing.T) {
	for _, path := range []string{"../../../testdata/abdada_fens.txt", "../../../testdata/lichess_2026_10_09_loss_fens.txt"} {
		positions, err := loadBenchPositions(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, position := range positions {
			parsed, err := ParseFEN(position.FEN)
			if err != nil {
				t.Fatal(err)
			}
			parsed.Board.CacheGetAllMoves = false
			parsed.Board.CacheGetAllAttackableMoves = false
			countLegalMoveNodes(t, parsed.Board, parsed.Active, parsed.Previous, 1)
		}
	}
}

func BenchmarkKingCheck(b *testing.B) {
	positions, err := loadBenchPositions("../../../testdata/abdada_fens.txt")
	if err != nil {
		b.Fatal(err)
	}
	boards := make([]*board.Board, 0, len(positions))
	for _, position := range positions {
		parsed, err := ParseFEN(position.FEN)
		if err != nil {
			b.Fatal(err)
		}
		parsed.Board.CacheGetAllAttackableMoves = false
		boards = append(boards, parsed.Board)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		boards[i%len(boards)].IsKingInCheck(color.Color(i % 2))
	}
}
