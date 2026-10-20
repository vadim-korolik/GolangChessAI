package ai

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/location"
)

func TestMateInOneSafetyRejectsKingWalkAndKeepsCheckingDefense(t *testing.T) {
	// 3R4/5k1p/4n3/1P1p1p2/6pK/1P1r4/8/8 w - - 2 55
	b := &board.Board{}
	b.LoadBoardFromText([]string{
		"   |   |   |   |   |   |   |   ",
		"   |   |   |   |   |   |   |   ",
		"   |   |   |   |B_R|   |W_P|   ",
		"W_K|B_P|   |   |   |   |   |   ",
		"   |   |B_P|   |B_P|   |W_P|   ",
		"   |   |   |B_N|   |   |   |   ",
		"B_P|   |B_K|   |   |   |   |   ",
		"   |   |   |   |W_R|   |   |   ",
	})
	b.SetFlag(board.FlagKingMoved, color.White, true)
	b.SetFlag(board.FlagKingMoved, color.Black, true)
	b.MovesSinceNoDraw = 2
	unsafe := location.Move{Start: location.NewLocation(3, 0), End: location.NewLocation(4, 0)}
	safe := location.Move{Start: location.NewLocation(7, 4), End: location.NewLocation(6, 4)}
	legal := b.GetAllMoves(color.White, nil)
	if !isMoveInList(unsafe, legal) || !isMoveInList(safe, legal) {
		t.Fatal("mate-safety fixture must allow both Kh5 and Rd7+")
	}
	if !moveAllowsMateInOne(b, nil, color.White, unsafe) {
		t.Fatal("Kh5 must allow Rh3#")
	}
	if moveAllowsMateInOne(b, nil, color.White, safe) {
		t.Fatal("Rd7+ must prevent the immediate mating reply")
	}
}

func TestMateInOneSafetyPrefersDeliveringMate(t *testing.T) {
	for _, tt := range []struct {
		name, placement string
		side            color.Color
		start, end      location.CoordinateType
	}{
		{"white", "3R3Q/5k1p/4n3/1P1p1pP1/6pK/1P1r4/8/8", color.White, 3, 4},
		{"black", "8/8/1p1R4/6Pk/1p1P1Pp1/4N3/5K1P/3r3q", color.Black, 4, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := boardFromFENPlacement(t, tt.placement)
			b.SetFlag(board.FlagKingMoved, color.White, true)
			b.SetFlag(board.FlagKingMoved, color.Black, true)
			b.MovesSinceNoDraw = 2
			unsafe := location.Move{Start: location.NewLocation(tt.start, 0), End: location.NewLocation(tt.end, 0)}
			if !isMoveInList(unsafe, b.GetAllMoves(tt.side, nil)) || !moveAllowsMateInOne(b, nil, tt.side, unsafe) {
				t.Fatal("king walk must be legal and permit an immediate mating reply")
			}
			hash := b.Hash()
			p := NewAIPlayer(tt.side, &ABDADA{NumThreads: 1})
			best := p.avoidImmediateMateMove(b, nil, &ScoredMove{Move: unsafe, Score: LossScore})
			child := b.Copy()
			last := board.MakeMove(&best.Move, child)
			if !child.IsInCheckmate(tt.side^1, last) || best.Score != WinScore {
				t.Fatalf("fallback must select and recognize immediate mate, got %s score=%d", best.Move, best.Score)
			}
			if b.Hash() != hash {
				t.Fatal("fallback must preserve the root board")
			}
		})
	}
}
