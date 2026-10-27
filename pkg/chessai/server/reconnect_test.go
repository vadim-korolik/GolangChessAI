package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/game"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/player/ai"
)

func reconnectTestClient(t *testing.T) (*Client, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var moves []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/stream/") {
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			return
		}
		mu.Lock()
		moves = append(moves, r.URL.Path)
		mu.Unlock()
		io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)
	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return &Client{BaseURL: base, APIKey: "test", HttpClient: srv.Client()}, func() []string { mu.Lock(); defer mu.Unlock(); return append([]string(nil), moves...) }
}

func reconnectTestGame(t *testing.T, side color.Color) *Lichess {
	t.Helper()
	p := randomAI(side)
	p.Opening = ai.OpeningNone
	p.Debug = false
	p.PrintInfo = false
	opponent := player.NewHumanPlayer(side ^ 1)
	var g *game.Game
	if side == color.White {
		g = game.NewGame(p, opponent)
	} else {
		g = game.NewGame(opponent, p)
	}
	l := &Lichess{GameID: "reconnect", Player: p, Game: g, GameEvents: make(chan GameEvent, 4)}
	t.Cleanup(func() { l.Mutex.Lock(); defer l.Mutex.Unlock(); l.resetGame() })
	return l
}

func TestDuplicateGameStartPreservesActiveGame(t *testing.T) {
	client, requests := reconnectTestClient(t)
	l := reconnectTestGame(t, color.Black)
	l.Client = client
	for _, uci := range []string{"g1f3", "d7d5"} {
		l.Game.PlayTurnMove(parseUCIMove(uci))
		l.movesApplied++
	}
	originalGame, originalPlayer, hash := l.Game, l.Player, l.Game.CurrentBoard.Hash()
	err := l.handleEvent(&Event{Type: EventTypeGameStart, Game: &Game{GameID: l.GameID, Color: "black", SecondsLeft: 180}})
	if err != nil {
		t.Fatal(err)
	}
	if l.Game != originalGame || l.Player != originalPlayer || l.movesApplied != 2 || l.Game.CurrentBoard.Hash() != hash || l.Game.GameStatus != game.Active {
		t.Fatal("duplicate gameStart must preserve the active game, player, board, and move count")
	}
	if len(requests()) != 0 {
		t.Fatal("duplicate gameStart must not post a move")
	}
}

func TestStaleGameFullCannotUndoPostedReply(t *testing.T) {
	l := reconnectTestGame(t, color.Black)
	for _, uci := range []string{"g1f3", "d7d5"} {
		l.Game.PlayTurnMove(parseUCIMove(uci))
		l.movesApplied++
	}
	original, hash := l.Game, l.Game.CurrentBoard.Hash()
	l.Mutex.Lock()
	err := l.handleGameFullLocked(&GameEvent{Status: "started", Moves: "g1f3", WhiteTimeMS: 180000, BlackTimeMS: 180000})
	l.Mutex.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 2 || l.Game.CurrentBoard.Hash() != hash || l.Game.CurrentTurnColor != color.White {
		t.Fatal("a snapshot older than our posted reply must not reset or advance the board")
	}
}

func TestReconnectPreservesSC7DT4VsReplyAndPlaysNextTurn(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/lichess_2026_10_19_reconnect_game.json")
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
	if len(replay) != 57 || replay[55].UCI != "a6b5" || replay[56].UCI != "c1b1" {
		t.Fatal("recorded reconnect fixture must retain the posted reply and next opponent move")
	}
	l := reconnectTestGame(t, color.Black)
	recorder := &recordingClient{}
	base, _ := url.Parse("http://test.local")
	l.Client = &Client{BaseURL: base, APIKey: "test", HttpClient: recorder}
	var moves []string
	for _, record := range replay {
		moves = append(moves, record.UCI)
	}
	for _, uci := range moves[:56] {
		l.Game.PlayTurnMove(parseUCIMove(uci))
		l.movesApplied++
	}
	original, hash := l.Game, l.Game.CurrentBoard.Hash()
	err = l.handleBoardUpdate(&GameEvent{Type: StateTypeGameFull, GameID: l.GameID, State: &GameEvent{
		Status: "started", Moves: strings.Join(moves[:55], " "), BlackTimeMS: 76190, BlackIncMS: 2000,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 56 || l.Game.CurrentBoard.Hash() != hash || len(recorder.urls) != 0 {
		t.Fatal("stale 55-ply snapshot discarded or repeated our acknowledged 56th-ply reply")
	}
	err = l.handleBoardUpdate(&GameEvent{
		Type: StateTypeGame, GameID: l.GameID, Status: "started", Moves: strings.Join(moves, " "),
		BlackTimeMS: 75570, BlackIncMS: 2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 58 || l.Game.CurrentTurnColor != color.White || len(recorder.urls) != 1 {
		t.Fatal("bot must preserve the game and post exactly one legal reply to the next opponent move")
	}
}

func TestPonderCanStopImmediatelyAfterStarting(t *testing.T) {
	l := reconnectTestGame(t, color.Black)
	parsed, err := analysis.ParseFEN("6kr/5p2/8/8/8/8/2P5/R5K1 w - - 0 20")
	if err != nil {
		t.Fatal(err)
	}
	l.Game.CurrentBoard = parsed.Board
	l.Game.CurrentTurnColor = parsed.Active
	l.Player.Algorithm = &ai.ABDADA{NumThreads: 1}
	l.Player.MaxSearchDepth = 64
	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			l.startPonder()
			l.stopPonder()
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("immediate ponder cancellation waited for the 60-second search budget")
	}
}

type reconnectHTTPFunc func(*http.Request) (*http.Response, error)

func (f reconnectHTTPFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestBoardStreamCancellationUnblocksPendingEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	base, _ := url.Parse("http://test.local")
	l := &Lichess{Client: &Client{BaseURL: base, APIKey: "test", HttpClient: reconnectHTTPFunc(func(*http.Request) (*http.Response, error) {
		close(ready)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("{\"type\":\"gameFull\",\"state\":{\"moves\":\"g1f3\"}}\n"))}, nil
	})}}
	events := make(chan GameEvent)
	done := make(chan error, 1)
	go func() { done <- l.StreamBoardUpdate(ctx, "reconnect", events) }()
	<-ready
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled stream returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancelled stream remained blocked sending an obsolete event")
	}
}
