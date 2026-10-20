# Mate-safety evaluation perspective, October 15, 2026

Status: validated locally; deployment pending. No Elo gain is established.

After search proposes a move permitting mate in one, the production player
examines safe alternatives. It previously evaluated each child using its own
color as the side to move. The child is the opponent's turn: terminal detection
therefore checked the wrong king. A mating alternative could receive an ordinary
static score and lose the comparison against a nonmating move. The one-line fix
evaluates the opponent's turn and negates the resulting score for our ranking.

A constructed regression extends the existing rook-ending mate-safety fixture:
`3R3Q/5k1p/4n3/1P1p1pP1/6pK/1P1r4/8/8 w - - 2 55`.
The legal proposal Kh5 permits Rh3#. The original fallback selects Rd7 with
score +980, missing the available immediate mate. After the fix, it selects
Qf6# with WinScore. A color-mirrored fixture selects Qf3#. Both verify terminal
mate, correct winning score, and preservation of the root board. These are
constructed tests, not newly observed live positions.

Focused mate-safety, quiescence, evaluation, and en passant checks pass in
0.042 seconds. The AI suite passes in 73.597 seconds with the two documented
timing-sensitive exclusions, TestAIBestMovesSame and
TestLazySMPFindsHangingKnight, plus existing in-test skips. Targeted mate-safety
race checks pass in 1.029 seconds. Production build uses GOAMD64=v3, the
existing cpu.pprof, and `-ldflags="-s -w"`. Optimized probes inject the legal
unsafe proposal through the real AIPlayer.GetBestMove path and verify Qf6#
and Qf3#, taking 1.177 and 0.860 ms respectively. The PGO profile is unchanged.

Seventeen recorded safety-override calls were replayed with full UCI history.
Each proposal and logged override must be uniquely legal, and the proposal
must still permit immediate mate. Two interleaved log entries require the
position two plies earlier: interrupted pondering can print after newer game
states arrive. The comparison uses the current optimized source and a Go
compiler overlay restoring only the old evaluation caller. All 17 selected
replies remain identical and prevent immediate mate. These calls include
pondering, so they are not all actual played fallback moves. This demonstrates
preservation on the sampled history, not an improved game result. Core-search
corpora were not rerun because those benchmarks bypass the changed outer guard.

Proof, exact positions, replays, test results, and hashes:
`testdata/strength_2026_10_15_mate_perspective_validation.json`.

## Separate queen-position diagnosis

The earlier o6lAdkjC-ply48 Rb5 horizon miss was examined before this fix.
Cold-hash single-threaded Stockfish depth 22 scores Rb5 at -511, while the
normal-budget Rb6 remains winning at +585. A depth-five matrix still selects
Rb5 without TT; disabling all forward pruning changes the chosen move but
does not establish a safe global remedy. Fresh forced-root scoring reduces
Rb5 from +1139 to +743 when all pruning is disabled, still optimistic at this
finite depth. This supports a shallow search/evaluation horizon rather than
a TT-only defect. No pruning or evaluation change is retained from this probe.
Exact referee lines and ablations are saved in
`testdata/lichess_2026_10_15_queen_horizon_diagnostics.json`.
The existing deeper normal-budget choice is winning; this diagnosis does not
justify a new tactical exception.
