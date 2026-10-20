# Central opening replies, October 15, 2026

Status: deployed October 15 at 00:37:14 America/Los_Angeles / 07:37:14 UTC.
The bot is online. Its first completed game is recorded below; the new
central opening replies have not appeared in fresh play.

Seventeen games from the previous deployment enter either 1.d4 d5 2.c4 Nf6
or 1.Nf3 d5 2.c4 Nf6. They comprise seven wins and ten losses against different
opponents; those counts do not establish causation. The recorded second move
uses only 86.781 microseconds in mv1fVMW2 and 374.584 microseconds in RY4YvAIT,
despite a 15-second clock allocation: fixed opening logic bypasses search.
The opening policy is unchanged by the deployed cache fix, and the current
regression reproduces Nf6. This uses prior-build positions to diagnose an
unchanged policy, not to claim fresh-game strength from the cache build.

Cold-hash, single-threaded Stockfish depth 22 prefers e6 in the Queen's Gambit
position (score -27 versus Nf6 -78) and d4 in the English position (score -2
versus Nf6 -74). The two-position referee loss is 61.5 cp before and zero for
the proposed replies. This small local opening result is not an Elo estimate.
The later e6-versus-Nxd5 positions were also checked; they do not support a
general forced-recapture rule. Search at the actual 15-second/eight-worker
budget still selects Nf6 in both Queen's Gambit runs, while both English runs
select d4. More clock alone does not resolve both positions.

The change extends the existing exact-position opening-response helper and
runs it before the blind scripts. It selects 2...e6 after 1.d4 d5 2.c4 and
2...d4 after 1.Nf3 d5 2.c4. Full board hash, castling flags, exact previous
move, and legal move membership must match. Existing Petroff and Scandinavian
responses use the same helper. The broad scripts, preference list, search,
evaluation, cache policy, and clock allocation are unchanged.

The new regression fails on the deployed source and passes after the patch.
It covers both positions with scripts disabled and with each script selected,
plus absent last-move context, changed rights, and board preservation.
Existing opening regressions pass in 0.864 seconds. The AI suite passes in
91.407 seconds with the two documented timing-sensitive tests explicitly
excluded (TestAIBestMovesSame and TestLazySMPFindsHangingKnight), and existing
in-test skips retained. No existing assertion was changed. The first race attempt was deferred by the live-game supervisor. After that
game ended, opening and cache race checks pass in 2.167 seconds. The optimized
real-player probe imports both exact FENs, validates all three script states,
returns the intended legal move without searching, preserves the board hash,
and takes 105-243 microseconds per reply. Production build and git diff --check
pass.

Exact FENs: `testdata/lichess_2026_10_15_central_opening_fens.txt`.
Proof and referee lines: `testdata/strength_2026_10_15_central_opening_validation.json`.

## Deployment and fresh cache-build game

The pending rapid game FZUllGuc ended in a White win by mate after 83 plies,
against a 1675-rated bot at 10+5. This was the first completed game after the
cache-pressure deployment; it predates the new opening responses and therefore
cannot validate them. Rapid rating refreshed to 1936. One expected win against
a weaker opponent does not isolate cache-policy strength or establish an Elo
gain.

Deployed only after authenticated playing status became idle. PID 1804913
runs SHA-256
`c66509fe05ee95e65e718746124ad9a33371654470b76734b8c603f6a35395ef`.
Installed and running hashes match; Lichess reports the bot online. All
30,113,777 pre-deployment log bytes were verified preserved through append
mode. Rollback: `/tmp/chess-oct15-main-before-central-openings` (the validated
cache-pressure build). The separate web server, PID 1931717, remains running.
The playing engine stays in Go; Stockfish is used only as referee.
The 2400 Elo objective remains unmet.

## First completed opening-build game

Om4tlrmF ended October 15 at 00:58:19 America/Los_Angeles in a Black mate win
against ElaChessBot, rated 1703, after 108 plies at 10+2. The game began
1.e4 c5 2.Nf3 Nc6 3.Bb5 Nf6, so neither new central opening reply triggered.
Rapid rating increased from 1936 to 1939. One expected win against a weaker
opponent does not isolate the opening changes or establish an Elo gain.
The exported game, including clocks, is preserved in
`testdata/lichess_2026_10_15_central_opening_games.json`.
