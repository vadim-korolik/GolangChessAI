package ai_test

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func TestABDADAPreservesBishopAgainstConnectedRookPassers(t *testing.T) {
	const fen = "3k4/7R/4p3/3b4/2p1BK2/8/3r4/8 w - - 4 65"
	scores := map[string]int{}
	for _, uci := range []string{"e4d5", "f4e3"} {
		parsed, err := analysis.ParseFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		parsed.Board.CacheGetAllMoves = false
		parsed.Board.CacheGetAllAttackableMoves = false
		algorithm := &ai.ABDADA{NumThreads: 1}
		player := ai.NewAIPlayer(parsed.Active, algorithm)
		player.TranspositionTableEnabled = true
		player.Debug = false
		player.PrintInfo = false
		move, err := analysis.MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, uci)
		if err != nil {
			t.Fatal(err)
		}
		scores[uci] = algorithm.ScoreRootMove(player, parsed.Board, move, 8).Score
	}
	t.Logf("Bxd5=%d Ke3=%d", scores["e4d5"], scores["f4e3"])
	if scores["f4e3"] <= scores["e4d5"] {
		t.Fatalf("Ke3 should preserve defense rather than create connected passers: %v", scores)
	}
}
