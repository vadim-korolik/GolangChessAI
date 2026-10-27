# Pawnless king activity, October 26, 2026

Status: the narrower forward-rank-only change is validated locally and deployed.
Production SHA-256:
`18f5032b7d01785219b0dd3ab4fa26f62b2333fa124057369701da5c7f030e32`.
No Elo improvement is established; the 2400 target remains unmet. Verified
ratings are rapid 1985, blitz 1815, bullet 1897.

## Recorded failure

In rapid game JsmSeUmZ, White played 65.Bxd5 with 226.59 seconds left and a
6000 ms search budget. After ...exd5, Black has connected c4/d5 passers in a
rook ending. Stockfish depth 22 preferred Ke3: unrestricted score -197 cp,
restricted Ke3 -209 cp, restricted Bxd5 -538 cp. The saved production build
repeated Bxd5 three times at the recorded 6000 ms budget and four workers.
The complete game is in `testdata/lichess_2026_10_26_rook_bishop_game.json`;
the exact position is in `testdata/lichess_pawnless_king_2026_10_26_fens.txt`.

The custom queenless king term gives White's f4 king 430 internal units
and Black's d8 king 60. After the trade, the complete static evaluator reads
White +3 cp despite Black's two connected passed pawns. Ke3 loses 110 units
of forward-rank reward, discouraging useful defense. Fresh depth-eight
forced-root searches scored Bxd5 -53 and Ke3 -79. Disabling LMR, null move,
futility, or razoring separately did not change those scores. Omitting the
custom king activity term reversed the ranking: Bxd5 -152, Ke3 -131.

## First experiment: omit the entire custom term

Keep the normal tapered king tables and passer terms. Omit the custom
queenless king activity bonus for both sides when exactly one side has
pawns. King roles then depend on the passers rather than advancement from
each king's starting rank. Preserve the term when both sides have pawns or
both sides are pawnless.

Removing only the pawnless defender's forward-rank bonus left Bxd5 and
Ke3 tied at -186. Removing that defender's entire custom bonus left both
moves tied at -235. Neither variant passed the unchanged new search
regression. The symmetric gate passes: Bxd5 -152, Ke3 -131. The static
connected-passer regression and its color mirror also pass.

The timed replay is only partially improved: candidate
`491138055205f9b5e7ee615a92768f8735222e8463d6082b2c254315fa204a5e`
chose Ke3 in two runs that reached depth seven and the six-second deadline.
One run completed depth eight in 5.855 seconds and still chose Bxd5.
The aggregate benchmark reports the majority move; inspect individual runs
before interpreting its zero missed-expected count as complete recovery.

## Validation

The AI suite passes in 50.904 seconds with `TestAIBestMovesSame` and
`TestLazySMPFindsHangingKnight` excluded. Two existing known-weakness tests
also skip themselves: `TestABDADACdTsyKE6AvoidsQuietBlunderWhenPawnBreakAvailable`
and `TestABDADACdTsyKE6AvoidsForcedMateWalk`.

Focused race checks pass in 132.309 seconds, covering Stockfish evaluation,
the new exact position, the existing king activation, Bxg6, rook attack,
knight retreat, and bishop defense. Existing tactical assertions remain
unchanged. The temporary component-printing test was archived outside the
repository. The PGO profile remains unchanged.

The paired corpus uses 3000 ms and four workers, alternating order with
fresh processes and common cold-hash Stockfish depth-18 scoring. Report
the original 94 positions separately from the four prior focus positions
and this new ending. Each child runs under the authenticated idle supervisor;
interrupted pairs are excluded and retried. Mean loss is capped at 1000 cp
per position, with mate-scale errors counted separately.

| Corpus | Capped mean before/after | Errors >=100 cp | Errors >=200 cp | Mate-scale errors |
| --- | --- | --- | --- | --- |
| Original 94 | 102.9 / 77.6 cp | 28 / 25 | 13 / 9 | 1 / 0 |
| All 99 | 105.6 / 88.1 cp | 31 / 28 | 15 / 11 | 1 / 1 |

The vVm9J7Ja ply-72 king move changed from ...f6 to the recorded ...Kf8
mistake. Common scoring prefers the candidate's move in this comparison:
247 cp loss before, 204 cp after. It remains a poor move, but the recorded
bad-move label alone does not show a regression against this baseline.

The added rqwLdsu1 ending initially changed from ...g5 to ...Qg6, walking
into mate. This blocked deployment. Three replay runs then reversed the
result: the old binary played ...Qg6 every time; the candidate played ...g5
every time. Both executables therefore exhibit unstable outcomes on this
position. Those grouped runs shared a process per binary. A second paired
run repeats all 13 changed corpus positions three times, using a fresh
process for every search and alternating order. This is a selected subset,
not a second full-corpus estimate.

The 39 repeat pairs scored 276.7 cp before and 128.1 cp after, with errors
at least 100 cp falling from 22 to 12 and errors at least 200 cp from 19 to
9. Mate-scale errors fell from four to three. The original queen ending,
gAacjwE9 ply 130, 9OVaO5TG, and the new JsmSeUmZ position improved in every
repeat. The gdKA8SFK improvement did not repeat. The rqwLdsu1 mate holdout
remained worse: one of three baseline searches entered mate, compared with
all three candidate searches. Grouped runs in one process had given the
opposite result, so they were insufficient to clear the concern.

Fresh single-worker forced-root comparisons at depths five and seven give
identical scores and node counts on rqwLdsu1 for both executables. At depth
five ...g5 scores -781 and ...Qg6 -526; at depth seven they score -863 and
-1408 respectively. This shows a serious horizon/timing weakness shared by
both builds. It does not erase the candidate's worse timed holdout results.
The entire-term removal has not been deployed.

## Narrower experiment: omit forward-rank rewards only

Keep `10*centerActivity` for both kings. Add `110*relRank` only when both
sides have pawns or both sides are pawnless. With pawns on exactly one side,
neither king receives the automatic forward-rank reward. This targets the
diagnosed mechanism while preserving centralization.

The static connected-passer and color-mirror regression passes, as does the
unchanged king-activation regression. Fresh forced-root depth-eight scores
are Bxd5 -142, Ke3 -109. At the recorded 6000 ms budget and four workers,
the candidate chose Ke3 in all three runs, completed depth eight every time,
and averaged 5.366 seconds. Candidate SHA-256:
`18f5032b7d01785219b0dd3ab4fa26f62b2333fa124057369701da5c7f030e32`.
The full AI gate passes in 50.283 seconds with the same two timing exclusions
and two existing in-test skips listed above. Focused race checks pass in
116.758 seconds.
The first AI attempt was excluded when XA6844Za started; no validation CPU
work continued during that live game. The production clock-safety build
won by mate on move 16, and the completed checks resumed afterward.
The fresh paired corpus improves all aggregate checks:

| Corpus | Capped mean before/after | Errors >=100 cp | Errors >=200 cp | Mate-scale errors |
| --- | --- | --- | --- | --- |
| Original 94 | 88.5 / 69.4 cp | 24 / 23 | 11 / 8 | 1 / 0 |
| All 99 | 102.1 / 70.2 cp | 28 / 25 | 14 / 9 | 2 / 0 |

These are matched comparisons within this new run. The same baseline binary
can choose different moves between runs, so absolute averages across the two
experiments are not directly comparable. No Elo gain follows from these local
measurements.

The rqwLdsu1 mate holdout then received four additional paired searches with
balanced before/after order and a fresh process per search. Both builds
entered mate in two of four searches and scored the same capped mean of
527.5 cp. The candidate's first full-corpus success on this position therefore
does not establish that the shared timing weakness is fixed.

Two other full-corpus positions increased by more than 100 cp:
IN3a4FSp ply 13 (38 to 171 cp loss) and hzqrcLk5 ply 43 (0 to 183 cp).
Four fresh paired searches per position produced seven errors at least
100 cp in eight baseline searches and eight in eight candidate searches.
The capped means were 154.1 and 177.0 cp on this selected subset. One baseline
search in hzqrcLk5 found the better move; the other seven baseline searches
matched the candidate's poorer moves. This residual short-budget weakness
remains visible rather than being excluded from the full-corpus result.

## Deployment and remaining limits

Deployed at 2026-10-26T21:53:26.208543-07:00 after authenticated idle confirmation.
Bot PID 3556086, installed `main`, and the validated candidate share the
`18f5032b...` hash above. Lichess reports the bot online. All 93,332,258 prior
log bytes are verified preserved, and the idle supervisor now follows the new
PID. Web server PID 1931717 remains untouched. The PGO profile remains
`fd489831db47b5aa5fbd41e581a2a56d876223aeb19b4c67073c57c46cbc88ab`.
Rollback binary: `/tmp/chess-oct26-main-before-pawnless-forward`.

Existing tactical assertions remain unchanged. The broader entire-term
removal was never deployed; centralization is preserved in production.
The new code improves the recorded ending and the measured local corpus,
but shared mate/horizon instability remains. The first completed post-build
game, ZGrXBCDu, was a rated 3+2 Black win against collider2 (1583), by mate
after 114 plies with 87.67 seconds left. Blitz rating moved 1813 to 1815.
The full export is in `testdata/lichess_2026_10_26_post_pawnless_games.json`.
This confirms continued runtime play on the new build; one win against a
lower-rated opponent does not isolate an evaluation strength or Elo gain.
Complete results, source patches, replay runs, and deployment evidence are in
`testdata/strength_2026_10_26_pawnless_king_validation.json`.
