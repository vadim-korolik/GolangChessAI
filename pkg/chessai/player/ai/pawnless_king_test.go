package ai

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
)

func TestStockfishClassicPawnlessKingDoesNotOffsetConnectedPassers(t *testing.T) {
	position := boardFromFENPlacement(t, "3k4/7R/8/3p4/2p2K2/8/3r4/8")
	mirrored := boardFromFENPlacement(t, "8/3R4/8/2P2k2/3P4/8/7r/3K4")
	white := evaluateStockfishClassicScore(position, color.White)
	black := evaluateStockfishClassicScore(mirrored, color.Black)
	// JsmSeUmZ: Bxd5 exd5 creates connected passers. A pawnless king on f4
	// must not erase the two-pawn deficit by advancing away from their paths.
	if white >= -100 {
		t.Fatalf("connected rook passers should remain an advantage: White=%d", white)
	}
	if white != black {
		t.Fatalf("color-mirrored pawnless defense differs: White=%d Black=%d", white, black)
	}
}
