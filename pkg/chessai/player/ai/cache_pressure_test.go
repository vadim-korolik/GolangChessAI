package ai

import (
	"runtime/debug"
	"testing"

	"github.com/Vadman97/GolangChessAI/pkg/chessai/color"
	"github.com/Vadman97/GolangChessAI/pkg/chessai/util"
)

func TestClearCachesReleasesOwnedABDADAWorkers(t *testing.T) {
	for _, tt := range []struct {
		name        string
		force       bool
		memoryLimit int64
		wantClear   bool
	}{
		{"forced", true, 1 << 62, true},
		{"memory-pressure", false, 1, true},
		{"below-budget", false, 1 << 62, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			algorithm := &ABDADA{NumThreads: 2}
			p := NewAIPlayer(color.Black, algorithm)
			algorithm.player = p
			algorithm.ensureRootWorkers(2)
			oldEvaluation, oldTT := p.evaluationMap, p.transpositionTable
			hash := util.BoardHash{1}
			oldEvaluation.Store(&hash, color.White, &evaluationPair{score: 42, whoMoves: color.White})
			oldWorker := algorithm.rootWorkerABs[0]
			oldSelect := algorithm.selectAB

			previousLimit := debug.SetMemoryLimit(tt.memoryLimit)
			defer debug.SetMemoryLimit(previousLimit)
			p.ClearCaches(tt.force)

			if tt.wantClear {
				if p.evaluationMap == oldEvaluation || p.transpositionTable == oldTT {
					t.Fatal("cache clearing must replace both parent maps")
				}
				if len(algorithm.rootWorkerABs) != 0 || algorithm.selectAB != nil || algorithm.workersFor != nil {
					t.Fatal("cleared cache must not remain retained by the persistent worker pool")
				}
				algorithm.ensureRootWorkers(2)
				if algorithm.rootWorkerABs[0] == oldWorker || algorithm.selectAB == oldSelect || algorithm.rootWorkerABs[0].player.evaluationMap != p.evaluationMap {
					t.Fatal("next search must rebuild workers against the current cache")
				}
			} else if p.evaluationMap != oldEvaluation || p.transpositionTable != oldTT || algorithm.rootWorkerABs[0] != oldWorker || algorithm.selectAB != oldSelect {
				t.Fatal("below-budget cleanup must retain useful caches and workers")
			}
		})
	}
}
