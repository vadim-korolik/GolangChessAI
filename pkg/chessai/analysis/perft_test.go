package analysis

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
)

func TestStandardLegalMoveCounts(t *testing.T) {
	// Standard perft positions cover castling, en passant, pins, checks, and promotions.
	positions := []struct {
		name, fen string
		nodes     int
	}{
		{"start", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 8902},
		{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 97862},
		{"rook-pawn", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 2812},
		{"promotion-castling", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 9467},
		{"promotion-check", "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8", 62379},
	}
	for _, position := range positions {
		t.Run(position.name, func(t *testing.T) {
			parsed, err := ParseFEN(position.fen)
			if err != nil {
				t.Fatal(err)
			}
			parsed.Board.CacheGetAllMoves = false
			parsed.Board.CacheGetAllAttackableMoves = false
			if got := countLegalMoveNodes(t, parsed.Board, parsed.Active, parsed.Previous, 3); got != position.nodes {
				t.Fatalf("depth-three nodes = %d, want %d", got, position.nodes)
			}
		})
	}
}

func countLegalMoveNodes(t *testing.T, b *board.Board, side color.Color, last *board.LastMove, depth int) int {
	assertKingCheckMatchesAttackMap(t, b)
	if depth == 0 {
		return 1
	}

	moves := b.GetAllMovesUnShuffled(side, last)
	nodes := 0
	for _, move := range *moves {
		child := b.Copy()
		previous := board.MakeMove(&move, child)
		nodes += countLegalMoveNodes(t, child, side^1, previous, depth-1)
	}
	return nodes
}
