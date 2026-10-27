package server

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/game_config"
)

func TestThinkTimeForClockReservesAvailableTimeBeforeIncrement(t *testing.T) {
	cases := []struct {
		name                      string
		clock, increment, maximum time.Duration
	}{
		{"recorded_flag", 1160 * time.Millisecond, time.Second, 580 * time.Millisecond},
		{"increment_exceeds_clock", 250 * time.Millisecond, 5 * time.Second, 125 * time.Millisecond},
		{"below_old_floor", 80 * time.Millisecond, 0, 40 * time.Millisecond},
		{"unknown_clock", 0, time.Second, time.Millisecond},
		{"expired_clock", -time.Second, time.Second, time.Millisecond},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := thinkTimeForClock(tc.clock, tc.increment, 122)
			if got <= 0 || got > tc.maximum {
				t.Fatalf("think time %s must stay positive and <= %s with %s available before %s increment", got, tc.maximum, tc.clock, tc.increment)
			}
		})
	}
}

func TestThinkTimeForPositionRecordedIncrementFlag(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/lichess_2026_10_26_increment_flag_game.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded struct{ Moves string }
	if err := json.Unmarshal(data, &recorded); err != nil {
		t.Fatal(err)
	}
	replay, err := analysis.ReplaySANMoves(recorded.Moves)
	if err != nil {
		t.Fatal(err)
	}
	if len(replay) != 245 || replay[244].UCI != "c5c6" {
		t.Fatal("recorded clock-loss fixture changed")
	}
	last := replay[244]
	parsed, err := analysis.ParseFEN(last.FENAfter)
	if err != nil {
		t.Fatal(err)
	}
	if !isCriticalSearchPosition(parsed.Board, color.Black) {
		t.Fatal("recorded rook ending must remain critical")
	}
	got := thinkTimeForPosition(1160*time.Millisecond, time.Second, 122, parsed.Board, color.Black)
	if got <= 0 || got > 580*time.Millisecond {
		t.Fatalf("recorded clock loss allocated %s with only 1.16s available", got)
	}
}

func TestCriticalClockBoostCannotSpendUnearnedIncrement(t *testing.T) {
	conf := game_config.Get()
	before := conf.AIMaxThinkTimeMs
	conf.AIMaxThinkTimeMs = 30000
	defer func() { conf.AIMaxThinkTimeMs = before }()
	parsed, err := analysis.ParseFEN("8/7k/4r3/1K4RP/4N1P1/5P2/8/8 b - g3 0 65")
	if err != nil {
		t.Fatal(err)
	}
	got := thinkTimeForPosition(9*time.Second, 20*time.Second, 65, parsed.Board, color.Black)
	if got <= 0 || got > 4500*time.Millisecond {
		t.Fatalf("critical boost exceeded available-clock safety cap: %s", got)
	}
}
