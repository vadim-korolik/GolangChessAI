package ai

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/location"
)

func TestContactAttackerPinnedToQueen(t *testing.T) {
	for _, tc := range []struct {
		name, placement string
		row, col        int
		want            bool
	}{
		{"rook file", "6kq/8/8/8/8/8/6Kb/7R", 1, 0, true},
		{"white rook pin", "7r/6kB/8/8/8/8/8/6KQ", 6, 0, true},
		{"rook rank", "6k1/8/8/8/8/8/Rb5q/6K1", 1, 6, true},
		{"bishop diagonal", "6k1/8/8/7q/8/5n2/8/3B2K1", 2, 2, true},
		{"bishop captures pinner", "6k1/8/8/7q/8/5b2/8/3B2K1", 2, 2, false},
		{"friendly slider", "6kq/8/8/8/8/8/6Kb/7r", 1, 0, false},
		{"wrong slider", "6kq/8/8/8/8/8/6Kb/7B", 1, 0, false},
		{"blocker toward queen", "6kq/8/8/8/8/7p/6Kb/7R", 1, 0, false},
		{"blocker toward pinner", "6kq/8/8/8/8/7b/6Kp/7R", 2, 0, false},
		{"opposing queen", "6kQ/8/8/8/8/8/6Kb/7R", 1, 0, false},
		{"no queen", "6k1/8/8/8/8/8/6Kb/7R", 1, 0, false},
		{"queen is not behind minor", "6kq/8/7R/8/8/8/6Kb/8", 1, 0, false},
		{"queen pinner", "6kq/8/8/8/8/8/6Kb/7Q", 1, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := kingDangerEvalForTest(boardFromFENPlacement(t, tc.placement))
			for _, p := range e.pieces {
				if p.row == tc.row && p.col == tc.col {
					if got := e.pinnedToQueen(p); got != tc.want {
						t.Fatalf("pinnedToQueen=%v, want %v", got, tc.want)
					}
					return
				}
			}
			t.Fatal("missing contact attacker")
		})
	}
}

func TestQueenPinReducesContactPressure(t *testing.T) {
	before := boardFromFENPlacement(t, "r1b1r1k1/3n1p1p/p6q/4p1p1/N3P1n1/2BB2P1/PP2NPKb/2RQ1R2")
	after := boardFromFENPlacement(t, "r1b1r1k1/3n1p1p/p6q/4p1p1/N3P1n1/2BB2P1/PP2NPKb/2RQ3R")
	unpinned := kingDangerEvalForTest(before).kingDanger(color.White)
	pinned := kingDangerEvalForTest(after).kingDanger(color.White)
	if pinned.mg >= unpinned.mg || pinned.eg >= unpinned.eg {
		t.Fatalf("Rh1 should reduce contact pressure: pinned=%+v unpinned=%+v", pinned, unpinned)
	}
	if w, b := evaluateStockfishClassicScore(after, color.White), evaluateStockfishClassicScore(after, color.Black); w != -b {
		t.Fatalf("evaluation symmetry: white=%d black=%d", w, b)
	}
}

func kingDangerEvalForTest(b *board.Board) *sfEval {
	e := &sfEval{b: b}
	var counts pieceTypeCounts
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			p := b.GetPiece(location.NewLocation(location.CoordinateType(row), location.CoordinateType(col)))
			if p == nil {
				continue
			}
			c, pt := p.GetColor(), p.GetPieceType()
			a := p.GetAttackableMoves(b)
			counts[c][pt]++
			e.pieceCount[c][pt]++
			sfSetBit(&e.occ[c], row, col)
			e.attackedBy2[c] |= e.attackedByAll[c] & a
			e.attackedByAll[c] |= a
			e.attackedBy[c][pt] |= a
			e.pieces = append(e.pieces, pcGeneric{row, col, c, pt, a})
		}
	}
	e.phase = endgamePhase(counts)
	e.occAll = e.occ[0] | e.occ[1]
	return e
}
