# Parallel fresh root verification, October 26, 2026

Status: rejected and reverted; the candidate was never deployed. Production
remains the deployed pawnless forward-rank fix, SHA-256
`18f5032b7d01785219b0dd3ab4fa26f62b2333fa124057369701da5c7f030e32`.
The over-2400 Elo objective remains unmet. No fresh games had completed on
that build at the start of this investigation.

## Evidence and scope

The inherited rqwLdsu1 ply-54 position remains unstable at 3000 ms and four
workers. Prior balanced fresh-process repeats entered mate in two of four
searches for both the pre-king-fix and deployed binaries. This is an existing
search weakness, not evidence against the new evaluation change.

A diagnostic overlay logs root iteration completion and fresh verification.
In its failing run, depth six took 2119 ms and timed out while verifying a
depth-eight candidate. Four same-depth fresh searches ran serially, followed
by the deeper tie-break. The engine discarded the incomplete verification
and correctly retained its last fully completed iteration, depth five, whose
Qg6 walks into mate. Two other traces finished depth seven and chose g5.

Earlier experiments that fixed sibling scout windows or accepted incomplete
verification were rejected by broader comparisons. This experiment preserves
dynamic root windows, every verification candidate and margin, isolated
fresh state, input-order tie handling, and the existing abort fallback.

## Candidate

Search independent fresh verification candidates concurrently, bounded by
the existing root thread count. Each search retains its own transposition
table and heuristics. Results occupy their original input indices. Join all
workers before reading results or returning, including after cancellation,
because the caller can mutate the live board immediately afterward.

Apply this only to the already-independent fresh searches in
`verifyParallelRootTie`, including its two deeper candidates. The shallow
rescue path reuses a sequential selection context and remains unchanged.
There is no new public API, setting, evaluation weight, or trust exception.
Candidate SHA-256:
`afb0305186793a25b6a8491edccfc698f05f8aca6abc188c57a03ed94a2445bb`.

## Initial proof and limits

The new fixed-depth test compares all four scores and moves with independent
serial searches, then reverses candidate order and checks every result again.
It also checks that verification preserves the live board. A separate test
cancels the owner after work starts and requires the batch to return without
clearing cancellation. These tests and the existing full-window root-scout
regression pass in 2.361 seconds; the new tests pass under the race detector
in 4.572 seconds.

Four alternating fresh-process pairs on the mate fixture still enter mate
twice in each build. The candidate's capped mean is 500.0 cp versus 527.5 cp
before, but this small sample does not establish a mate-avoidance improvement.
The candidate searches approximately 212k-219k nodes/second there, versus
163k-193k before.

Three alternating fresh-process pairs on the newly fixed JsmSeUmZ ending
use its actual 6000 ms/four-worker budget. Both builds choose Ke3 every time
and complete depth eight. Candidate throughput is 385k-394k nodes/second,
versus 237k-241k before. The candidate uses the full six seconds and still
returns depth eight; more processed nodes do not by themselves establish
stronger play.

The complete AI gate passes in 40.573 seconds with `TestAIBestMovesSame`
and `TestLazySMPFindsHangingKnight` excluded. Two existing known-weakness
tests skip themselves: `TestABDADACdTsyKE6AvoidsQuietBlunderWhenPawnBreakAvailable`
and `TestABDADACdTsyKE6AvoidsForcedMateWalk`. Relevant race checks pass in
93.746 seconds, including cancellation, the new ordered batch, full-window
root scouting, and the existing king, Bxg6, rook, knight, and bishop regressions.
All existing tactical assertions remain unchanged.

The completed 99-position production-budget comparison uses
3000 ms, four workers, fresh processes, alternating before/after order, and
common cold-hash Stockfish depth-18 scoring. The original 94 positions are
reported separately. Every child uses the authenticated idle supervisor;
an arriving live game excludes and retries an unfinished pair.

## Playing gate and rejection

| Corpus | Capped mean before/after | Errors >=100 cp | Errors >=200 cp | Mate-scale errors |
| --- | --- | --- | --- | --- |
| Original 94 | 83.2 / 84.9 cp | 25 / 24 | 10 / 9 | 0 / 1 |
| All 99 | 86.7 / 97.6 cp | 29 / 27 | 12 / 12 | 0 / 2 |

Candidate median throughput rises 21.0%. It completes a higher depth in
27 positions and a lower depth in six. The playing gate nevertheless worsens,
so neither the throughput increase nor passing correctness/race checks is
treated as a practical strength gain.

The new original-holdout mate is 33e5jNAp ply 69, changing safe Rc3 to Rb6.
Four additional balanced fresh-process pairs do not repeat that mate in
either build: all eight searches choose safe Rb7 instead. The known rqwLdsu1
mate also changes from safe Bf5 to Qg6 in the full corpus; the earlier four
paired replays had shown two mate failures in each build. These are unstable
timed outcomes rather than demonstrated deterministic score corruption.

The second critical holdout, a73hZRGs ply 42, changes Re6 (131 cp loss) to
Nb6 (391 cp loss). In four further pairs the baseline always chooses Re6;
the candidate chooses Re6 three times and Ra7 once (422 cp loss). Across the
eight critical replay pairs, capped mean is 65.5 cp before and 101.9 cp after.
This does not establish a useful playing improvement, even though the initial
mate regression does not repeat.

The independent-search helper and its tests were removed from the worktree.
The exact pre-experiment `abdada.go` was restored. A production-style rebuild
matches both installed and running executables byte-for-byte:
`18f5032b7d01785219b0dd3ab4fa26f62b2333fa124057369701da5c7f030e32`.
The source patch and test source remain preserved in
`testdata/strength_2026_10_26_parallel_verification_validation.json` together
with all corpus, replay, race, and trace evidence. The PGO profile is unchanged.

## Fresh runtime evidence

Game ZGrXBCDu interrupted the candidate's second side of pair index 55.
Both sides of that unfinished pair were excluded, and the runner resumed
after the game from 55 completed checkpoints. The deployed king-fix build
won this rated 3+2 game as Black against collider2 (1583), by mate after
114 plies with 87.67 seconds left. Blitz rating rose from 1813 to 1815;
rapid remains 1985. The full post-king-build export is preserved separately
in `testdata/lichess_2026_10_26_post_pawnless_games.json`.
This is functional runtime evidence for the retained build, not an isolated
Elo estimate. The parallel candidate never played a live game. The 2400 target
remains unmet, and shared search/horizon instability remains a follow-up.
