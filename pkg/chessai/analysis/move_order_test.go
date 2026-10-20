package analysis

import "testing"

func TestGetAllMovesOrdersCapturesWithTheirScores(t *testing.T) {
	// CdTsyKE6: a pawn capture must precede a knight capture and a queen capture of pawns.
	parsed, err := ParseFEN("r1bq1rk1/pp1n2bp/2p1pp1p/3pP3/3P4/2N1QN2/PPP2PPP/2KR1B1R w - - 0 10")
	if err != nil {
		t.Fatal(err)
	}
	moves := *parsed.Board.GetAllMoves(parsed.Active, parsed.Previous)
	want := []string{"e5f6", "c3d5", "e3h6"}
	if len(moves) < len(want) {
		t.Fatalf("got %d legal moves, need at least %d", len(moves), len(want))
	}
	for i, expected := range want {
		if got := MoveToUCI(moves[i]); got != expected {
			t.Fatalf("move %d = %s, want %s", i, got, expected)
		}
	}
}
