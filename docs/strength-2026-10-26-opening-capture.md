# Opening fallback capture guard, October 26, 2026

Status: validated locally and deployed. The over-2400 Elo objective
remains unmet. No isolated Elo gain is established.

## Fresh-game failure and cause

In ZGrXBCDu, the first game completed on the pawnless king-fix build, Black
answered 1.d4 d5 2.e4 with ...Nf6. The logged search depth was zero. White
could then gain space with e5. Cold-hash Stockfish depth 16 preferred ...dxe4:
the comparison measured 151 cp loss for ...Nf6. Common depth-18 scoring
measured 138 cp loss. The game was eventually won; the defect evidence is
the position and policy trace, rather than the result.

The scripted book already abandons a move that ignores a more profitable
legal capture, using its existing 100 cp SEE threshold. The fallback
`earlyOpeningPreference` checks only whether its proposed move hangs
material. It can therefore restore the same quiet development after the book
rejects it, or ignore a capture when book mode was already abandoned.
Here the fallback's first choice was ...Nf6. Search itself chose ...dxe4
in all three 3000 ms/four-worker probes.

The four-line change applies `bookMoveMissesFreeCapture` to fallback
preferences too. It preserves the existing threshold and defers to search;
it does not force a capture or add an opening-specific reply. Existing
complete-position responses retain their priority, including Petroff
counterplay and the Scandinavian and central responses.

## Proof

New regressions fail before and pass after for Blackmar and the analogous
White material opportunity in the Englund. Both positions are tested with book off
and both book choices. They establish that quiet development passes the
separate hanging-material gate, that the existing capture guard recognizes
the pawn, and that the full player policy reaches the search result without
mutating the board. All six cases fail on production source and pass after.

The focused opening gate passes in 1.111 seconds. The full AI suite passes
in 50.667 seconds, excluding `TestAIBestMovesSame` and
`TestLazySMPFindsHangingKnight`. Existing known-weakness tests
`TestABDADACdTsyKE6AvoidsQuietBlunderWhenPawnBreakAvailable` and
`TestABDADACdTsyKE6AvoidsForcedMateWalk` also skip themselves. Focused opening
race checks pass in 2.139 seconds. Existing tactical assertions are unchanged.

The policy corpus replays exact move histories from 44 recorded games,
covering 26 unique early prefixes. Each prefix runs with book off and both
book choices, at 500 ms and 3000 ms. Four additional standard gambit controls
bring the total to 30 prefixes and 180 policy cases. Every before/after
search starts a fresh process, and order alternates. This calls the complete
`AIPlayer.GetBestMove`, including opening policy. All work runs under the
authenticated idle supervisor. Common cold-hash Stockfish depth-18 scores
are cached only by identical FEN and forced move to avoid duplicate work.

| Cases per budget | Budget | Mean loss before/after | Errors >=100 cp before/after |
| --- | --- | --- | --- |
| Recorded prefixes and book states: 78 | 500 ms | 14.91 / 9.60 cp | 3 / 0 |
| Recorded prefixes and book states: 78 | 3000 ms | 14.91 / 9.60 cp | 3 / 0 |
| Recorded cases plus controls: 90 | 500 ms | 26.66 / 15.26 cp | 9 / 3 |
| Recorded cases plus controls: 90 | 3000 ms | 26.52 / 15.92 cp | 9 / 3 |

No individual paired case worsens. There are no errors at least 200 cp or
mate-scale errors in either build. The recorded Blackmar position changes
from ...Nf6 to ...dxe4 in all six book-state/budget cases. The Englund control
changes from Nf3 to dxe5 in all six. Short search still misses the best
...Nxe4 in the Alekhine pawn-offer control, although its chosen alternatives
improve against the old preference. This guard does not solve that search
weakness or establish an Elo estimate.

## Build and scope

Release candidate SHA-256:
`b114885617ba38a9180e3c8f73fce7626e46a73dafefa340755d367188c47e03`.
It uses `GOAMD64=v3`, the unchanged `cpu.pprof`, and `-ldflags="-s -w"`.
Search and evaluation are unchanged. The unrelated parallel-verification
experiment remains reverted. Full evidence, probe/runner source, per-case
results, and regression output are preserved in
`testdata/strength_2026_10_26_preference_capture_validation.json`.
Fresh post-deployment games remain necessary to assess practical strength.

Deployed at 2026-10-26T23:07:35.035282-07:00 after authenticated idle confirmation.
Bot PID 3711170, installed `main`, and the validated release share the
`b1148856...` hash above. Lichess reports the bot online. All 93,484,363 prior
log bytes are verified preserved, and the idle supervisor follows the new PID.
Rollback: `/tmp/chess-oct26-main-before-preference-capture`.

The saved web PID 1931717 had already disappeared before deployment. The
first preflight stopped before any changes. Replacement web PID 3671326 was
already serving port 3005; it retains its prior `18f5032b...` executable and
returns HTTP 200 after the bot deployment. The web process was not restarted.
Current verified ratings remain rapid 1985, blitz 1815, bullet 1897.
