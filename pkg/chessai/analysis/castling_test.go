package analysis

import "testing"

func TestCastlingChecksKingPathAndRook(t *testing.T) {
	cases := []struct {
		name, fen, move string
		legal           bool
	}{
		{"GWTjIExF-attacked-b8", "r3k1nr/pb1p1R1p/n1pBp2q/1p3PN1/3P4/P7/1PP2P1P/R2QKB2 b Qkq - 4 15", "e8c8", true},
		{"white-attacked-b1", "1r2k3/8/8/8/8/8/8/R3K3 w Q - 0 1", "e1c1", true},
		{"white-attacked-h1", "4k2r/8/8/8/8/8/8/4K2R w K - 0 1", "e1g1", true},
		{"black-attacked-h8", "4k2r/8/8/8/8/8/8/4K2R b k - 0 1", "e8g8", true},
		{"in-check", "4r1k1/8/8/8/8/8/8/R3K3 w Q - 0 1", "e1c1", false},
		{"attacked-d1", "3r2k1/8/8/8/8/8/8/R3K3 w Q - 0 1", "e1c1", false},
		{"attacked-c1", "2r3k1/8/8/8/8/8/8/R3K3 w Q - 0 1", "e1c1", false},
		{"attacked-f1", "5rk1/8/8/8/8/8/8/4K2R w K - 0 1", "e1g1", false},
		{"attacked-g1", "k5r1/8/8/8/8/8/8/4K2R w K - 0 1", "e1g1", false},
		{"occupied-b1", "4k3/8/8/8/8/8/8/RN2K3 w Q - 0 1", "e1c1", false},
		{"rook-blocks-f1", "4k3/8/8/8/8/8/8/4KR1R w K - 0 1", "e1g1", false},
		{"rook-blocks-d1", "4k3/8/8/8/8/8/8/R2RK3 w Q - 0 1", "e1c1", false},
		{"king-off-home-rank", "4k3/8/8/8/8/8/4K2R/8 w K - 0 1", "e2g2", false},
		{"king-off-home-file", "4k3/8/8/8/8/8/8/R2K4 w Q - 0 1", "d1b1", false},
		{"missing-rook", "4k3/8/8/8/8/8/8/4K3 w K - 0 1", "e1g1", false},
		{"wrong-color-rook", "4k3/8/8/8/8/8/8/r3K3 w Q - 0 1", "e1c1", false},
		{"wrong-corner-piece", "4k3/8/8/8/8/8/8/N3K3 w Q - 0 1", "e1c1", false},
		{"lost-rights", "4k3/8/8/8/8/8/8/R3K3 w - - 0 1", "e1c1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseFEN(tc.fen)
			if err != nil {
				t.Fatal(err)
			}
			_, err = MatchUCIMove(parsed.Board, parsed.Active, parsed.Previous, tc.move)
			if (err == nil) != tc.legal {
				t.Fatalf("%s: legal=%t, want %t", tc.move, err == nil, tc.legal)
			}
		})
	}
}

func TestReplayGWTjIExFCastlesPastAttackedB8(t *testing.T) {
	moves := "d2d4 c7c6 g1f3 e7e6 b1c3 d8a5 e2e4 f8b4 c1d2 a5h5 g2g4 h5g4 h1g1 g4h5 g1g7 f7f6 a2a3 b4c3 d2c3 b7b5 c3b4 b8a6 b4d6 f6f5 e4f5 h5h6 f3g5 c8b7 g7f7 e8c8"
	states, err := ReplayUCIMoves(moves)
	if err != nil {
		t.Fatal(err)
	}
	want := "2kr2nr/pb1p1R1p/n1pBp2q/1p3PN1/3P4/P7/1PP2P1P/R2QKB2 w Q - 5 16"
	if got := states[len(states)-1].FENAfter; got != want {
		t.Fatalf("castled position = %s, want %s", got, want)
	}
}
