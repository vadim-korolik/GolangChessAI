# Available-clock safety, October 26, 2026

Status: validated and deployed. The 2400 Elo objective remains active and
unmet; current rapid rating is 1983, blitz 1813, bullet 1897. No isolated Elo
gain is claimed from this change.

## Observed failure

In 3+1 blitz game Fz4rRXOm, Black reached a rook ending after 245 recorded
plies. The last board state left 1.16 seconds before Black's next move.
The allocator assigned a full second, because its exhausted reserve still
added the entire future one-second increment. The prior turn consumed
1.0083 seconds thinking and approximately 0.1717 seconds of other overhead.
The next search took 1.0066 seconds. The last post returned HTTP 200, but the
finished export contains only 245 plies and records Black losing on time.
The existing budget left too little delivery margin even without a transport
failure. The full game is preserved in
`testdata/lichess_2026_10_26_increment_flag_game.json`.

Both ordinary and critical final budgets now cap thinking at half the
currently available clock. A one-millisecond minimum prevents a missing or
expired clock from setting MaxThinkTime to zero, which disables the deadline.
The recorded 1.16-second case now allocates at most 580 ms. The same bound
also applies after a critical-position boost or a large configured cap.
Search and evaluation source remain unchanged from the prior production build.

## Proof and limits

New regressions fail on the old source for the recorded clock, an increment
larger than the bank, the old 50 ms floor near zero, missing/expired clocks,
and a critical boost with a large configured cap. The exact recorded 245-ply
replay also fails before and passes after. Existing healthy-clock and tactical
allocation assertions pass unchanged. Focused clock tests pass in 0.006 seconds;
the full server suite passes under the race detector in 4.573 seconds.

A differential replay covers 43 fresh standard games and 2,175 bot turns.
All 2,070 turns with at least eight seconds available retain identical budgets.
Only five played low-clock turns change, all near the end of Fz4rRXOm. The
flagged next turn is separately covered by the exact regression because it
never entered the finished move history. This validates budget preservation;
it does not prove a different result in the original game or a future Elo gain.
Full evidence, source patch, test output, and replay are in
`testdata/strength_2026_10_26_clock_safety_validation.json`.

The prior move-delivery fix now also has real recovery evidence: UNpKsgsg
suffered two connection resets while posting Re7 and g3. Both selected moves
were resubmitted successfully three seconds later, and the bot kept responding.
The game later ended as a chess loss by mate, rather than an abandoned clock
loss. This is functional runtime evidence, not an isolated strength estimate.

## Deployment

Deployed at 2026-10-26T20:40:58.583509-07:00 after authenticated idle confirmation.
Bot PID 3368784, installed executable, and production build match SHA-256
`b8da417d2a5ed96ea25160a4c01b67c6bbbae5d9b9641689b10f5bd03df31c34`.
Lichess reports the bot online. All 93,296,523 prior log bytes are verified
preserved. Web server PID 1931717 remains untouched, and the PGO profile is
unchanged. Rollback: `/tmp/chess-oct26-main-before-clock-safety`.
The idle supervisor now reads credentials from PID 3368784; old temporary
runners with PID 2161165 must be refreshed before reuse.

All contact-interposition and discovered-check search experiments were rejected
and reverted. A production rebuild of restored source matched the previous
live `5937b371...` executable exactly before adding this clock-only fix.
Their negative results remain documented separately in
`docs/strength-2026-10-25-discovered-check.md`.

Fresh post-build loss diagnosis identified fifteen Stockfish depth-16 mistakes
in WJMX9IM7, fu9Ek8Fg, and JsmSeUmZ. Exact positions are preserved in
`testdata/lichess_2026_10_26_loss_fens.txt`; they remain search/evaluation leads,
not fixes claimed by this clock change.
