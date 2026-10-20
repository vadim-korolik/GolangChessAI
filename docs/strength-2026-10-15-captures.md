# Selective quiescence generation, October 15, 2026

Status: deployed for measured throughput. Timed strength results are mixed;
fresh-game validation is pending, and no Elo improvement is established.

The preceding search profile attributes roughly 43% of sampled CPU to full
move generation and another 3.45% to duplicate terminal evaluation. Quiescence
outside check generated every legal quiet move before discarding it, then
static evaluation repeated legal-move detection. This experiment removes that
work without changing evaluation weights or search pruning.

`Board.GetCapturesAndPromotions` reuses the existing piece generators and
legality checks in their original unsorted order. A scoped, generator-local
flag skips quiet destinations except promotions, bypasses the full-move cache,
and preserves legal en passant. The flag is not copied into child boards.
Check evasions and randomized ordering still use full move generation. An
empty tactical list triggers ordinary legal-move detection so quiet positions
remain distinct from stalemate. After terminal and draw checks, quiescence
uses the existing cached static evaluation directly, including tempo and the
path-dependent fifty-move term. Other search algorithms retain their existing
terminal decisions through the shared boolean legal-move argument.

New regressions compare the exact ordered tactical list with filtered full
legal generation across the 94-position holdout, 42 recent loss positions,
en passant fixtures, checking passers, quiet and capture underpromotions,
pins, king captures, castling, and illegal discovered-check en passant.
They exercise cache-enabled and cache-disabled boards and verify board,
king-location, cached full-list, and subsequent uncached generation
preservation. Evaluation regressions cover both colors, cached scores,
halfmove counters 0/12/99/100, repetition, insufficient material, and
checkmate precedence at a draw boundary. Existing assertions are unchanged.

Focused evaluation tests pass in 0.038 seconds; board and analysis tests pass
in 0.005 and 0.044 seconds. The AI suite passes in 75.191 seconds with the two
documented timing-sensitive exclusions, `TestAIBestMovesSame` and
`TestLazySMPFindsHangingKnight`, plus existing in-test skips. Targeted race
checks pass in 250.948 seconds for AI and 1.169 seconds for analysis, including
the eight-worker promotion-mate root test. Every CPU job used the live-game
idle supervisor. Production build uses GOAMD64=v3 and the existing cpu.pprof
with `-ldflags="-s -w"`; the production PGO profile is unchanged.

Four alternating batches of eight independent single-worker searches per
build retain forced-Qe3 depth-five score +356 and forced-Kh5 depth-seven mate
score -1000000002 in every run. Mean node rates rise from 86,748 to 112,030
(+29.1%) for Qe3 and 118,659 to 147,319 (+24.2%) for Kh5. A separate depth-three
single-worker comparison across all 94 holdout positions preserves every
move, score, depth, and considered-node count exactly: 3,632,986 nodes for
each build. Summed reported search time drops from 37,107 to 28,693 ms
(22.7%), equivalent to 29.3% higher aggregate node throughput. These are local
throughput measurements, not evidence of an Elo gain.

The candidate SHA-256 is
`de2b65f4e5a4dbbe29ca8c1be8c114ba93b43015571c6e43e07ceb8deeecb698`.
The deployed opening-build baseline is
`c66509fe05ee95e65e718746124ad9a33371654470b76734b8c603f6a35395ef`.
Raw comparison artifacts use `/tmp/chess-oct15-capture-` prefixes.
The 2400 Elo objective remains unmet.

## Timed strength checks and regression investigation

One run per position used the same production binaries, budgets, worker counts,
and common cold-hash, single-threaded Stockfish depth-18 referee scores.
Preferred referee moves are explicitly assigned zero loss. Loss means below
cap each position at 1000 cp; mate-scale counts use 10000 cp.

| Metric | Opening-build baseline | Selective generation |
|---|---:|---:|
| 94 positions, 1200 ms/four workers: capped mean cp | 123.9 | 114.5 |
| Losses >=100 / >=200 cp | 32 / 17 | 28 / 15 |
| Mate-scale losses | 3 | 3 |
| 16 positions, 6000 ms/eight workers: capped mean cp | 157.7 | 212.4 |
| Losses >=100 / >=200 cp | 7 / 4 | 7 / 4 |
| Mate-scale losses | 1 | 1 |

Neither set introduces a new mate-scale loss. Total considered nodes increase
from 23,202,614 to 28,776,610 over the holdout and from 13,642,897 to 17,185,172
over the tactical set. Holdout mean depth increases from 5.79 to 6.10; tactical
mean depth remains 5.94. These mixed strength results do not establish a
broad strength improvement.

The largest holdout regression is o6lAdkjC-ply48: Rf8 at 89 cp loss becomes
Rb5 at 924 cp loss after a deeper iteration completes. Three further
alternating 1200 ms/four-worker repeats reproduce that move difference in
every run. Both builds select Rb5 with identical +1185 score and TT-prune
count in a fixed-depth-five single-worker check. At the production 2500 ms
floor/eight workers, both select Rf8 in two of two runs. At 15000 ms/eight
workers, both select Rb6; the candidate reaches depth seven in 12.744 seconds,
while baseline times out at depth six. The bad shallow horizon remains; the
optimization exposes it sooner rather than repairing it.

The other >=100 cp holdout regression, gAacjwE9-ply44, chooses Rf4 in one of
three further candidate repeats at 1200 ms, versus none of three baseline
repeats. Both builds select Kf7 in two of two production-floor/eight-worker
runs. Its intermittent short-budget weakness remains.

The tactical mean worsens mainly because mv1fVMW2-ply58 switches from Nxe4
to Re4 (1237 cp loss). In three further alternating 6000 ms/eight-worker
repeats, each build chooses Nxe4 once and Re4 twice. At the actual logged
11229 ms budget, both choose Nxe4 in all three repeats. The candidate averages
2,523,630 considered nodes versus 2,253,816 and reaches depth eight in one
repeat; baseline reaches depth seven in all three. This confirms the earlier
six-second outcome is unstable and does not reproduce a normal-budget
critical-reply regression. It does not fix the inherited six-second weakness.

Retained as a throughput improvement because ordered generation, fixed-depth
scores/moves/nodes, and checked production-budget replies are preserved.
Fresh-game validation remains required. Complete local proof is preserved in
`testdata/strength_2026_10_15_selective_captures_validation.json`.

## Deployment

Deployed at 2026-10-15T01:28:46.407783-07:00 after authenticated playing status reported
no active game. PID 1870344 runs the candidate SHA-256 recorded above; installed
and running executable hashes match. Lichess reports the bot online.
All 30,259,390 preceding log bytes were verified preserved through
append mode. The idle-game supervisor now checks the new PID. Rollback is
`/tmp/chess-oct15-main-before-selective-captures`. The separate web server,
PID 1931717, remains running. Stockfish remains the referee.
No completed game from this deployment has been assessed.
