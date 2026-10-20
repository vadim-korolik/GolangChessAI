package ai_test

import (
	"slices"
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/location"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestQuiescePreservesCachedLegalMovesAndBoard(t *testing.T) {
	parsed, err := analysis.ParseFEN("4k3/8/2q5/3P4/8/8/8/R3K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	b := parsed.Board
	b.CacheGetAllMoves = true
	cached := b.GetAllMovesUnShuffled(parsed.Active, parsed.Previous)
	before := append([]location.Move(nil), (*cached)...)
	hash := b.Hash()
	p := ai.NewAIPlayer(parsed.Active, &ai.ABDADA{NumThreads: 1})
	if score := p.Quiesce(b, ai.NegInf, ai.PosInf, parsed.Active, parsed.Previous); score <= 0 {
		t.Fatalf("quiescence search must see dxc6 winning the queen, got score=%d", score)
	}
	if !slices.Equal(before, *cached) || b.Hash() != hash {
		t.Fatal("quiescence search must preserve the cached legal moves and root board")
	}
}
