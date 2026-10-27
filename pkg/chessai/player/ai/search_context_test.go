package ai

import (
	"context"
	"testing"
	"time"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/board"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
)

type abortResetSearch struct {
	started, release chan struct{}
	observed         chan bool
}

func (*abortResetSearch) GetName() string { return "abort reset probe" }

func (a *abortResetSearch) GetBestMove(p *AIPlayer, b *board.Board, previous *board.LastMove) *ScoredMove {
	close(a.started)
	<-a.release
	p.ResetAbort()
	a.observed <- p.isAborted()
	return &ScoredMove{Move: (*b.GetAllMoves(p.PlayerColor, previous))[0]}
}

func TestSearchContextCancellationSurvivesAbortReset(t *testing.T) {
	for _, beforeStart := range []bool{true, false} {
		name := "during startup"
		if beforeStart {
			name = "before startup"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			algorithm := &abortResetSearch{make(chan struct{}), make(chan struct{}), make(chan bool, 1)}
			p := NewAIPlayer(color.White, algorithm)
			p.Opening, p.TurnCount = OpeningNone, 12
			p.Debug, p.PrintInfo = false, false
			b := &board.Board{}
			b.ResetDefault()
			if beforeStart {
				cancel()
			}
			done := make(chan struct{})
			go func() {
				p.GetBestMoveContext(ctx, b, nil, nil)
				close(done)
			}()
			<-algorithm.started
			cancel()
			close(algorithm.release)
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("cancelled search did not finish")
			}
			if !<-algorithm.observed {
				t.Fatal("per-search abort reset discarded owner cancellation")
			}
			if p.isAborted() {
				t.Fatal("finished search retained a cancelled context")
			}
		})
	}
}
