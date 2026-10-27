# Qxh4 deflection and blocked bishop pressure, October 25, 2026

Status: all experiments rejected and reverted. At the start, the bot ran the
validated pin-pressure and move-delivery build, SHA-256
`5937b371dc34cd2aefa442b64aad4218ea30d3fa0ecae66ddf2694d6ecabb3f1`.
No fresh games had completed at the start of these experiments. Rapid was
1989; the 2400 Elo objective remains active and unmet.

## Exact remaining miss

Current production still chooses Bh3 in qkmR1kyN-ply30 at its actual 5272 ms,
four-worker budget. It also still chooses Rde8 in SYmzLcwj-ply50 at 9859 ms;
the former g5 position chooses Re6 in this replay. These are current-build
reproductions of older exact positions, not fresh-game strength evidence.

Cold-hash Stockfish depth 22 prefers Qxh4. Restricted searches score Bh3 +30
and Qxh4 +310; unrestricted search scores Qxh4 +332. One line is Qxh4 gxh4
Be2+ Qg5 Rxg5+ hxg5 Nxc2. Restricted and unrestricted finite-depth scores
can differ; no exact universal score claim follows.

Two independent mechanisms explain the wrong ranking. First, Bh3 changes
static Black score from +28 to +430 even though the queen attacks no king-ring
square. White's Ng2 interposes a defended knight on the bishop's only contact
square, but the old contact bonus still contributes about 274 cp. Excluding
contact targets blocked by defended opposing pieces changes the post-Ng2
score from +332 to +58. Actual checks remain eligible even when a friendly
piece protects the checked king's square. All attack maps and weights remain
unchanged.

Second, LMR reduces Be2+ by one ply at a depth-four node, move index six,
after Qxh4 gxh4. Moving the bishop uncovers the g8 rook's check against Kg1.
The reduction leaves the queen sacrifice at +80; no-LMR finds +251. The
candidate searches an uncovered check at full depth when the moved piece
itself does not attack the king. Ordinary direct checks retain their existing
reduction policy. Root verification does not rescue this line initially:
Qxh4's +80 lies below the two moves sent to the deeper tie-break.

Fresh single-worker depth-six controls demonstrate both contributions:

| Change | Bh3 | Qxh4 |
|---|---:|---:|
| Baseline | 419 | 80 |
| Evaluation only | 150 | 80 |
| Search only | 419 | 251 |
| Both | 150 | 251 |

The combined candidate selects Qxh4 in all three actual-budget replays.
The exact search regression and static interposition regression pass, as do
unchanged Bxg6, bishop-defense, rook-attack, and knight-retreat assertions.
Focused checks take 17.449 seconds. Full AI/race gates and an alternating,
fresh-process 98-position three-second/four-worker corpus are queued before
retention. The original 94 positions remain separately reported holdout data.

Requiring queen ring support alone did not fix the timed miss. Applying the
old total-pawn shelter gate to minors broke the retained Rh1 evaluation case;
counting only front shelter broke the unchanged bishop-defense assertion.
All three are rejected; no assertion was weakened. Full sources, raw referee
lines, ablations, traces, tests, and validation output are preserved in
`testdata/strength_2026_10_25_discovered_check_validation.json`.

Full AI validation passes in 74.878 seconds with TestAIBestMovesSame and
TestLazySMPFindsHangingKnight explicitly excluded, plus existing in-test skips.
Relevant AI race checks pass in 31.275 seconds. A production rebuild matches
the immutable candidate byte-for-byte. All four depth-six controls complete
with the scores above. The 98-position paired corpus is now running; no
retention or deployment follows until its broader results are evaluated.

## Combined candidate fails broader quality gate

The complete 98-position three-second comparison worsens capped mean loss
101.29 to 110.34 cp, >=100 cp errors 28 to 30, and >=200 cp errors 13 to 15.
Mate-scale errors remain two. The original 94-position holdout worsens
90.43 to 97.57 cp, >=100 errors 25 to 27, and >=200 errors 11 to 12; mate-scale
errors remain one. Eleven moves differ. Qxh4 saves 295 cp, but the older
Rbf5 position regresses 835 cp, the former g5 position regresses 287 cp, the
Bh2/Rde8 position regresses 223 cp, and a knight defense loses 183 cp after
finishing two fewer depths. Other changes recover some loss but do not make
the combined candidate a broad improvement. It will not be deployed as is.
Separate evaluation-only and search-only builds are now compared on the five
material positions to identify which mechanism causes the regressions. The
live bot remains on the previously validated build; no new Elo claim follows.

## Narrow defended-minor candidate

Component replays preserve Rf8 in baseline and search-only builds, while the
broad evaluation rule prefers Rb5. The old g5 position likewise changes Re6
to Rb8 only in builds with that evaluation change. In the knight-depth case,
all four builds choose Rc1/depth four in the controls, so the earlier isolated
paired depth difference is timing-sensitive. The Bh2/Rde8 position changes to
Rde8 under either component, requiring broader scoring rather than a causal
claim from one pair.

The next evaluation variant limits interposition to defended knights and
bishops: the equal-material shield actually seen with Ng2. Pressure against
protected pawns and major pieces keeps production behavior. This avoids
removing other load-bearing attack terms. All existing focused regressions
pass in 16.847 seconds. Three actual-budget replays still select Qxh4.
At three seconds, the five material controls now select Rf8, Rc1, Re6, Qxh4,
and Bh2, respectively. Rc1 matches every build's repeated control, while
Rf8/Re6/Bh2 restore the production selections. Full gates and a new paired
98-position corpus are required before retaining this narrower variant.

## Evaluation guard rejected; search change isolated

The narrow 98-position comparison is nearly flat overall: capped mean loss
106.93 to 106.52 cp, >=100 errors 30 to 29, >=200 errors 15 to 13, and two
mate-scale errors in both. Original 94 positions worsen 93.94 to 99.02 cp,
>=100 errors 26 to 27, >=200 errors twelve in both, one mate-scale error in
both. The largest new loss is vVm9J7Ja-ply30: d5 to Bc3, 671 cp, after losing
one completed depth. In three alternating component repeats, baseline and
search-only find d5 3/3. Evaluation-only finds Bc3 3/3, and the combined build
finds it 2/3. Thus this is not dismissed as the isolated knight-depth noise.

A more conservative support check preserves pressure on twice-attacked minor
blockers, but still chooses Bc3 in the material control and changes Bh2 to
Rde8 again. It is rejected before another broad run. All interposition
changes are restored byte-for-byte to the deployed evaluation source. Their
experimental static tests are archived with the evidence, not left failing
or skipped in the worktree. No prior user tactical assertion changed.

The remaining experiment is the independent discovered-check search change.
Its test now isolates the accepted search mechanism: Qxh4's continuation must
score at least +150, versus baseline +80 and candidate +251. It deliberately
does not claim to prefer Qxh4 over Bh3: the unresolved contact evaluation
still calls Bh3 +419 at this depth. Full AI/race gates and a separate paired
98-position corpus will decide whether this search-only change helps playing
choices broadly. No experimental build has been deployed.

## Final rejection and exact restoration, October 26

Search-only's completed 98-position comparison worsens capped mean loss
105.30 to 106.82 cp. >=100 errors fall 30 to 29, >=200 errors rise 14 to 15,
and mate-scale errors remain two. On the original 94 positions, capped mean
worsens 92.27 to 93.37 cp, >=100 errors remain 26, >=200 errors rise eleven
to twelve, and mate-scale errors remain one. Qxh4's continuation is evaluated
more accurately, but this does not produce a validated broad playing gain.
The original Bh3 choice remains, and other timed choices offset any benefit.

The search-only change is rejected along with every contact-interposition
variant. ABDADA and evaluation source are restored exactly to the saved
production bytes; the new unretained search regression is archived in the
validation artifact and removed from the worktree. All pre-existing tactical
assertions remain intact. A production-style restored rebuild exactly matches
SHA-256 `5937b371dc34cd2aefa442b64aad4218ea30d3fa0ecae66ddf2694d6ecabb3f1`.
No experimental binary was deployed. A separate fresh-game clock-budget fix
was subsequently validated and deployed, as recorded in
`docs/strength-2026-10-26-clock-safety.md`.
