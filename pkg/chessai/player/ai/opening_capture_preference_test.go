package ai

import (
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/location"
)

func TestOpeningPreferenceDefersToProfitableCapture(t *testing.T) {
	for _, tc := range []struct {
		name, capture string
		line          []string
		side          color.Color
	}{
		{"Blackmar", "d5e4", []string{"d2d4", "d7d5", "e2e4"}, color.Black},
		{"Englund", "d4e5", []string{"d2d4", "e7e5"}, color.White},
	} {
		for _, opening := range []struct {
			name  string
			index int
		}{{"none", OpeningNone}, {"first", 0}, {"second", 1}} {
			t.Run(tc.name+"/"+opening.name, func(t *testing.T) {
				b := &board.Board{}
				b.ResetDefault()
				var previous *board.LastMove
				for i, uci := range tc.line {
					previous = gateApply(t, b, previous, color.Color(i%2), uci)
				}
				capture := location.Move{Start: gateSquare(t, tc.capture[:2]), End: gateSquare(t, tc.capture[2:])}
				p := NewAIPlayer(tc.side, &fixedOpeningRegressionAlgorithm{move: capture})
				p.TurnCount, p.Opening = 1, opening.index
				p.Debug, p.PrintInfo = false, false
				quiet := openingPreferenceMoves(tc.side, p.TurnCount)[0]
				if !p.bookMoveMissesFreeCapture(b, previous, quiet) {
					t.Fatal("existing capture gate must recognize the missed pawn")
				}
				if p.preferenceMoveHangsMaterial(b, quiet) {
					t.Fatal("quiet development should pass the separate hanging-material gate")
				}
				if got := p.earlyOpeningPreference(b, previous); got != nil {
					t.Errorf("preference ignored profitable capture %s: got %s", tc.capture, got.UCIString())
				}
				before := b.Hash()
				got := p.GetBestMove(b, previous, nil)
				if !got.Start.Equals(capture.Start) || !got.End.Equals(capture.End) {
					t.Errorf("expected search capture %s, got %s", tc.capture, got.UCIString())
				}
				if p.Opening != OpeningNone || b.Hash() != before {
					t.Fatal("capture fallback must leave book mode and preserve the board")
				}
			})
		}
	}
}
