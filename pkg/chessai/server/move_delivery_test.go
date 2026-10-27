package server

import (
	"encoding/json"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/analysis"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
)

type deliveryClient struct {
	stream  string
	bodies  []string
	queries []string
	mu      sync.Mutex
	codes   []int
	failure error
	moves   []string
}

func (c *deliveryClient) Do(r *http.Request) (*http.Response, error) {
	if r.Method == http.MethodGet {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(c.stream))}, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queries = append(c.queries, r.URL.Query().Get("offeringDraw"))
	c.moves = append(c.moves, strings.TrimPrefix(r.URL.Path, "/api/bot/game/reconnect/move/"))
	if len(c.moves) == 1 && c.failure != nil {
		return nil, c.failure
	}
	status := http.StatusOK
	if len(c.codes) > 0 {
		status = c.codes[0]
		c.codes = c.codes[1:]
	}
	body := `{"ok":true}`
	if len(c.bodies) > 0 {
		body = c.bodies[0]
		c.bodies = c.bodies[1:]
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func (c *deliveryClient) posted() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.moves...)
}

func deliveryGame(t *testing.T, c *deliveryClient) *Lichess {
	t.Helper()
	l := reconnectTestGame(t, color.Black)
	base, _ := url.Parse("http://test.local")
	l.Client = &Client{BaseURL: base, APIKey: "test", HttpClient: c}
	return l
}

func deliveryState(moves string, full bool) *GameEvent {
	state := &GameEvent{Type: StateTypeGame, GameID: "reconnect", Moves: moves, Status: "started", WhiteTimeMS: 131820, BlackTimeMS: 131820}
	if full {
		return &GameEvent{Type: StateTypeGameFull, GameID: "reconnect", State: state}
	}
	return state
}

func TestMoveDeliveryRetriesSameMoveAfterGatewayFailure(t *testing.T) {
	c := &deliveryClient{codes: []int{http.StatusBadGateway, http.StatusOK}}
	l := deliveryGame(t, c)
	original := l.Game
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original {
		t.Fatal("HTTP 502 discarded the active game")
	}
	hash := l.Game.CurrentBoard.Hash()
	if l.movesApplied != 1 || len(c.posted()) != 1 {
		t.Fatal("unconfirmed reply must not be acknowledged")
	}
	if err := l.handleBoardUpdate(deliveryState("g1f3", true)); err != nil {
		t.Fatal(err)
	}
	moves := c.posted()
	if len(moves) != 2 || moves[0] != moves[1] {
		t.Fatalf("retry changed selected move: %v", moves)
	}
	if l.Game != original || l.Game.CurrentBoard.Hash() != hash || l.movesApplied != 2 {
		t.Fatal("retry changed board or failed to acknowledge reply")
	}
	if err := l.handleBoardUpdate(deliveryState("g1f3", true)); err != nil {
		t.Fatal(err)
	}
	if len(c.posted()) != 2 {
		t.Fatal("stale snapshot repeated an acknowledged reply")
	}
}

func TestMoveDeliveryAcceptsServerConfirmationAfterLostResponse(t *testing.T) {
	for _, full := range []bool{false, true} {
		t.Run(map[bool]string{false: "gameState", true: "gameFull"}[full], func(t *testing.T) {
			c := &deliveryClient{failure: io.ErrUnexpectedEOF}
			l := deliveryGame(t, c)
			original := l.Game
			if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
				t.Fatal(err)
			}
			if l.Game != original {
				t.Fatal("lost response discarded the active game")
			}
			selected := c.posted()[0]
			hash := l.Game.CurrentBoard.Hash()
			history := "g1f3 " + selected
			if err := l.handleBoardUpdate(deliveryState(history, full)); err != nil {
				t.Fatal(err)
			}
			if l.movesApplied != 2 || len(c.posted()) != 1 || l.Game.CurrentBoard.Hash() != hash {
				t.Fatal("confirmed reply was reapplied or reposted")
			}
			if err := l.handleBoardUpdate(deliveryState(history+" e2e4", false)); err != nil {
				t.Fatal(err)
			}
			if l.Game != original || l.movesApplied != 4 || len(c.posted()) != 2 {
				t.Fatal("bot failed to reply exactly once to next opponent move")
			}
		})
	}
}

func TestMoveDeliveryConfirmsReplyAndOpponentMoveTogether(t *testing.T) {
	c := &deliveryClient{codes: []int{http.StatusServiceUnavailable, http.StatusOK}}
	l := deliveryGame(t, c)
	original := l.Game
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original {
		t.Fatal("HTTP 503 discarded the active game")
	}
	history := "g1f3 " + c.posted()[0] + " e2e4"
	if err := l.handleBoardUpdate(deliveryState(history, true)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 4 || len(c.posted()) != 2 {
		t.Fatal("combined confirmation replay duplicated or lost a move")
	}
}

func TestMoveDeliveryStillDropsDefinitivelyRejectedBoard(t *testing.T) {
	c := &deliveryClient{codes: []int{http.StatusBadRequest}}
	l := deliveryGame(t, c)
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if l.Game != nil || l.Player != nil {
		t.Fatal("definitive illegal-move rejection must still discard desynced board")
	}
}

func TestMoveDeliveryHTTPErrorClassification(t *testing.T) {
	for _, code := range []int{400, 401, 403, 429, 500, 502, 503, 504} {
		c := &deliveryClient{codes: []int{code}}
		l := deliveryGame(t, c)
		original := l.Game
		if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
			t.Fatal(err)
		}
		retry := code == 429 || code >= 500
		if (l.Game == original) != retry {
			t.Errorf("HTTP %d: active game preserved=%v, want %v", code, l.Game == original, retry)
		}
	}
}

func TestMoveDeliveryRecoversFirstWhiteMove(t *testing.T) {
	c := &deliveryClient{codes: []int{502, 200}}
	base, _ := url.Parse("http://test.local")
	l := &Lichess{Client: &Client{BaseURL: base, APIKey: "test", HttpClient: c}, GameEvents: make(chan GameEvent, 4)}
	t.Cleanup(func() { l.Mutex.Lock(); defer l.Mutex.Unlock(); l.resetGame() })
	if err := l.handleEvent(&Event{Type: EventTypeGameStart, Game: &Game{GameID: "reconnect", Color: "white", SecondsLeft: 180}}); err != nil {
		t.Fatal(err)
	}
	if l.Game == nil {
		t.Fatal("first move's HTTP 502 discarded game before initial snapshot")
	}
	original, hash := l.Game, l.Game.CurrentBoard.Hash()
	if err := l.handleBoardUpdate(deliveryState("", true)); err != nil {
		t.Fatal(err)
	}
	moves := c.posted()
	if l.Game != original || l.movesApplied != 1 || l.Game.CurrentBoard.Hash() != hash || len(moves) != 2 || moves[0] != moves[1] {
		t.Fatal("first move recovery reset game or chose a different opening move")
	}
}

func TestMoveDeliveryDoesNotTrustRetryTurnRace(t *testing.T) {
	c := &deliveryClient{codes: []int{502, 400}, bodies: []string{"gateway error", `{"error":"Not your turn, or game already over"}`}}
	l := deliveryGame(t, c)
	original := l.Game
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if err := l.handleBoardUpdate(deliveryState("g1f3", true)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 1 {
		t.Fatal("retry's turn race must await server confirmation")
	}
	selected := c.posted()[0]
	if err := l.handleBoardUpdate(deliveryState("g1f3 "+selected, true)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 2 || len(c.posted()) != 2 {
		t.Fatal("server confirmation after retry race was lost")
	}
}

func TestMoveDeliveryRequestsFreshSnapshotWithoutOpponentEvent(t *testing.T) {
	c := &deliveryClient{codes: []int{502, 200}, stream: "{\"type\":\"gameFull\",\"state\":{\"moves\":\"g1f3\",\"status\":\"started\"}}\n"}
	l := deliveryGame(t, c)
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	select {
	case event := <-l.GameEvents:
		if err := l.handleBoardUpdate(&event); err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("unacknowledged move never requested a fresh board snapshot")
	}
	moves := c.posted()
	if l.Game == nil || l.movesApplied != 2 || len(moves) != 2 || moves[0] != moves[1] {
		t.Fatal("fresh snapshot did not retry the existing reply")
	}
}

func TestMoveDeliveryPreservesDrawOfferOnRetry(t *testing.T) {
	c := &deliveryClient{codes: []int{502, 200}}
	l := deliveryGame(t, c)
	l.offerDrawNextMove = true
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if err := l.handleBoardUpdate(deliveryState("g1f3", true)); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queries) != 2 || c.queries[0] != "true" || c.queries[1] != "true" {
		t.Fatalf("draw offer missing from actual query or retry: %v", c.queries)
	}
}

func TestMoveDeliveryRecordedGSWCClockLoss(t *testing.T) {
	data, err := os.ReadFile("../../../testdata/lichess_2026_10_25_clock_loss_game.json")
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
	if len(replay) != 42 || replay[40].UCI != "a1b2" || replay[41].UCI != "e7c6" {
		t.Fatal("clock-loss fixture changed")
	}
	c := &deliveryClient{codes: []int{502, 200}}
	l := reconnectTestGame(t, color.White)
	base, _ := url.Parse("http://test.local")
	l.Client = &Client{BaseURL: base, APIKey: "test", HttpClient: c}
	var history []string
	for i, r := range replay {
		history = append(history, r.UCI)
		if i < 41 {
			l.Game.PlayTurnMove(parseUCIMove(r.UCI))
			l.movesApplied++
		}
	}
	original := l.Game
	if err := l.handleBoardUpdate(deliveryState(strings.Join(history, " "), false)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 42 {
		t.Fatal("HTTP 502 abandoned the exact recorded position")
	}
	hash := l.Game.CurrentBoard.Hash()
	if err := l.handleBoardUpdate(deliveryState(strings.Join(history[:41], " "), true)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || len(c.posted()) != 1 || l.Game.CurrentBoard.Hash() != hash {
		t.Fatal("older snapshot reset the pending reply")
	}
	if err := l.handleBoardUpdate(deliveryState(strings.Join(history, " "), true)); err != nil {
		t.Fatal(err)
	}
	moves := c.posted()
	if l.Game != original || l.movesApplied != 43 || l.Game.CurrentBoard.Hash() != hash || len(moves) != 2 || moves[0] != moves[1] {
		t.Fatal("recorded clock-loss recovery failed to retry the same reply")
	}
}

type brokenDeliveryBody struct{ closed bool }

func (b *brokenDeliveryBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (b *brokenDeliveryBody) Close() error             { b.closed = true; return nil }

func TestMoveDeliveryKeepsGameAfterPartialResponse(t *testing.T) {
	c := &deliveryClient{}
	l := deliveryGame(t, c)
	original := l.Game
	body := &brokenDeliveryBody{}
	l.Client.HttpClient = reconnectHTTPFunc(func(r *http.Request) (*http.Response, error) {
		response, err := c.Do(r)
		if err == nil && r.Method == http.MethodPost && len(c.posted()) == 1 {
			response.Body = body
		}
		return response, err
	})
	l.offerDrawNextMove = true
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	if !body.closed || l.Game != original || l.movesApplied != 1 {
		t.Fatal("partial response must close body and preserve unconfirmed move")
	}
	if err := l.handleBoardUpdate(deliveryState("g1f3 "+c.posted()[0], true)); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 2 || l.offerDrawNextMove {
		t.Fatal("server confirmation must commit reply and consume its draw offer")
	}
}

func TestMoveDeliveryIgnoresUnrelatedEventsWhilePending(t *testing.T) {
	c := &deliveryClient{codes: []int{502}}
	l := deliveryGame(t, c)
	original := l.Game
	if err := l.handleBoardUpdate(deliveryState("g1f3", false)); err != nil {
		t.Fatal(err)
	}
	event := deliveryState("g1f3 a7a6", false)
	event.Type = "chatLine"
	if err := l.handleBoardUpdate(event); err != nil {
		t.Fatal(err)
	}
	if l.Game != original || l.movesApplied != 1 || len(c.posted()) != 1 {
		t.Fatal("unrelated event changed pending move state")
	}
}
