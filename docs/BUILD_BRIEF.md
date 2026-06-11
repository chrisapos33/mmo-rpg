# BUILD BRIEF — mmo-rpg (v1 MVP)

> **For: Claude Code (or equivalent coding agent) with direct access to the repo.**
> This document defines the v1 scope, architecture, aesthetic, and implementation
> order. Anything not in this file is OUT of v1 — see `VISION.md` for future scope.
> Do not build VISION items now.

---

## LOCKED DECISIONS (read first — these override anything below that conflicts)

These were decided after the repo inventory. They take precedence over the existing code
and the original brief wording.

**Decision 1 — Taxonomy: 5 dimensions + Trust as a META-score.**
The scoring model is FIVE dimensions: **Output/Cadence, Craft/Quality, Influence/Reach,
Collaboration, Range**. Trust is NOT a sixth peer dimension — it is a meta-score computed
from the confidence of the underlying evidence (see §4).
- The existing DB (`user_signal_scores`) and AI generator use the OLD six:
  Builder/Thinker/Executor/Collaborator/Specialist/**Trusted**. This is a mismatch to be
  migrated, NOT a thing to build on.
- **Drop `trusted` as a peer column** (it becomes the computed meta-score).
- **Keep the 7 existing classes** (Architect, Artisan, Pathfinder, Sage, Operator,
  Sentinel, Artificer) as flavor — they are MAPPED FROM the dimensions, not replaced.

**Decision 2 — Scoring model: ledger + decay/caps/normalize.**
Keep the existing `signal_events` table as the raw weighted-evidence **ledger** (good
design — `final_points = base × weight × confidence` already encodes quality + trust
weighting). But the engine, when aggregating, MUST apply **recency decay + per-source
caps** and **normalize to 0–100** (percentile, with a seeded reference distribution for
cold-start). **Do NOT expose cumulative raw points as the rank** — raw additive points
with no decay/cap is the "only-goes-up rewards volume" failure mode we are avoiding.

**Sequencing note:** the scoring engine (step 2) is a PURE module with fixtures and does
NOT touch the DB — so it is written directly against the new 5+Trust taxonomy now. The
schema migration (rename/restructure columns, drop `trusted`, update `domain/signal`, the
dimension mapping in `signal_service`, and the AI prompt) happens as ONE change set in the
persistence step, because this decision cuts across all of them.

**Decision 3 — Decay is PER-DIMENSION, not global.**
A single global half-life is wrong. Decay must vary by what the dimension measures:
- **Output/Cadence** — fast decay (it measures liveness; old activity should fade).
- **Influence/Reach** — minimal or NO decay. Dependents/reach are a *present-tense
  snapshot*, not a timestamped event. A library written 2 years ago that 5k repos import
  *today* has not lost value because the commit is old. Heavy decay here wrongly punishes
  durable work.
- **Craft/Quality (longevity)** — slow decay.
Do not apply one half-life across all dimensions.

**Decision 4 — Trust must NOT be dilutable. [Resolved & validated — accumulation model]**
A plain weighted-average confidence is wrong: adding weak signals (or even counted
signals weaker than the current mean, e.g. reviews-given at 0.72 added to a 0.92 build)
can *lower* Trust even when the developer did MORE verified work. That breaks the core
"connect GitHub → raise Trust" loop.
**Implemented model (locked):** keep a `minConfidence = 0.70` floor (signals below it
contribute to percentile dimension scores but are invisible to Trust), then
`Trust = strength / (strength + 70)` where `strength = Σ(points × confidence)` over
signals with confidence ≥ 0.70. This is **monotonic by construction**
(`dTrust/dstrength = 70/(strength+70)² > 0`): adding any verified signal can only raise
Trust, never lower it. Do not revert to a mean.
- *Known property, deferred tuning:* Trust can saturate from a single dominant source
  (e.g. one library with 8k verified stars → Trust ~0.99, above a broadly-active OSS
  contributor at ~0.93). Acceptable if Trust = "volume of verified evidence"; revisit
  only if Trust should instead reward *breadth* of sources.

---

## 0. STEP ZERO — Inventory before you build

Before writing or changing ANY code, scan the existing repo and report back:

- The full directory tree of `backend/` and `frontend/`.
- For the backend: list every file under `internal/` (repository, service, handler,
  api, ai, config, migrations). Summarize what each service and handler currently does.
- The existing DB schema: read all migration files and output the current tables,
  columns, and relationships.
- The frontend: framework, routing setup, existing pages/components, styling approach
  (CSS files, any UI library), and how it talks to the backend.
- `docker-compose.yml`, `Makefile`, `.env.example`: what services, env vars, and
  commands already exist.

**Output a short written inventory, then STOP and wait for confirmation before
implementing.** The point is to build ON TOP of what exists, not regenerate it.

Known facts already established (verify against reality):
- Backend: Go + `sqlx` + Postgres (`lib/pq`), migrations via `goose` (embedded FS,
  run on boot), JWT auth, Anthropic AI client gated behind a `MockAI` config flag.
- Layering: `repository → service → handler → router` (clean, keep it).
- Repos: User, CV, Profile, GitHub, Signal.
- Services: Auth (JWT), Onboarding (CV upload + AI profile), Signal, GitHub (OAuth).
- Handlers: Profile, Explore, Evidence.
- Frontend: TypeScript / React.

**Minor cleanup to flag (do not auto-fix without asking):** the Go module path is
`github.com/chrisapos3/mmo-rpg` but the repo is `chrisapos33/mmo-rpg` (double 3).
Harmless locally; matters only if the module ever needs to be `go get`-able.

---

## 1. What this product IS (v1)

A professional-identity tool for **developers only** (v1 is dev-only — this is a
deliberate constraint, not a limitation). A developer connects their GitHub account;
a scoring engine derives a multi-dimensional, **verification-aware** score from real
GitHub activity; an AI layer translates those scores into an MMORPG-style character
build (class, subclass, headline, flavor text); the result is shown as a high-fantasy
character profile with a procedurally generated emblem; and an Explore page lets users
discover other "hunters" filtered by class.

That is the entire v1. Four pages, one scoring engine, one AI layer, one emblem
generator.

### The core principle that governs the whole scoring engine

> **Rank by what OTHER people did in relation to your work. Anything you did alone in
> your own repos is liveness/context with a capped weight — never a ranker.**

Anything fully under the user's control (commit count, number of repos, lines of code,
"I know 12 languages") is trivially gameable. Anything that required a *third party* to
act on the user's work (a maintainer merged their PR, someone depends on their package,
someone reviewed their code) is expensive to fake. Verification IS third-party
attribution. This single principle drives both the scoring and the anti-gaming design.

---

## 2. Scope — IN vs OUT

### IN (build these)
- GitHub OAuth (log in **as the user**, use the user's token / rate limit).
- GitHub ingestion layer (REST + GraphQL v4 `contributionsCollection`).
- Scoring engine producing 5 dimensions + a Trust meta-score (isolated, testable module).
- AI generation layer: scores → class / subclass / headline / flavor text
  (single structured LLM call; respect the existing `MockAI` flag).
- Deterministic emblem generator (same build → same emblem).
- Persistence: users, builds, cached scores/signals, evidence.
- Four frontend pages: Landing, Auth/OAuth callback, Character/Profile, Explore.
- High-fantasy MMORPG aesthetic (see §5).

### OUT (do NOT build — these live in `VISION.md`)
- Anything company/recruiter-facing (the two-sided marketplace).
- Social feed, posts, activity stream, points-for-engagement.
- Open-source collaboration hub ("seeking contributors" posts).
- Support for non-developer professions (designers, marketers, PMs).
- Figurative AI-generated character art (Midjourney / DALL·E / Stable Diffusion).
- Public shareable profile URLs (nice, easy later — but not v1).

---

## 3. The CV vs GitHub relationship (important — resolves the two onboarding paths)

The existing code has **two** onboarding paths: `OnboardingService` (CV upload → AI
profile) and `GitHubService` (OAuth → signals). These are NOT redundant. They are the
two ends of the **trust spectrum**, and together they form the core gameplay loop:

- **CV = self-reported → a temporary, LOW-confidence build.** A first guess at
  class/dimensions. Gets the user a character immediately, before connecting anything.
- **GitHub connect = verified → RAISES confidence.** Validates and/or corrects the
  build using inspectable, attributable evidence.

The **Trust meta-score** is literally: *what fraction of the build rests on verifiable
(GitHub) evidence vs self-reported (CV) claims.* "Connect GitHub to make your character
real" is the loop.

**Decision for v1:** keep both paths, but the scoring engine and the demo are
**GitHub-centric**. CV is a low-confidence starting point, NOT an equal scoring source.
Do not invest in deep CV parsing; invest in GitHub.

---

## 4. The 5 dimensions + Trust (scoring engine spec)

> Taxonomy and aggregation model are fixed in **LOCKED DECISIONS** above (5 dimensions +
> Trust-as-meta; ledger + decay/caps/normalize). This section is the detail.

Build the scoring engine as an **isolated, independently testable module** (pure
functions: raw ingested data in → scores out, no DB or HTTP coupling). This is the heart
of the product and the thing future signals plug into without a rewrite.

Each dimension outputs 0–100. Apply the §1 principle throughout.

**Output / Cadence** — Distinct active days over a rolling window (~12–18 months) with
recency decay. NOT raw commit count. Active days count *only while the touched repos
carry external validation* (stars / external contributors / dependents); commits to a
solo, zero-signal repo ≈ 0. This is **table stakes, not a ranker** — answers "are they
alive?", not "are they good?". Cap it.

**Craft / Quality** — Proxies only, and be honest they're proxies: presence/ratio of
tests, CI config (Actions workflows), **depth of review the user's PRs receive before
merge** (GraphQL review threads — hard to fake, involves others), and **longevity**
(repos maintained over years, commits answering issues — time is expensive to forge).
README/docs = weak signal, low weight.
**LOCKED — validation gate (resolved):** tests/CI/longevity are *self-controlled* signals
(you can add a CI file and a tests/ folder to a solo repo with zero third-party
involvement). They must carry low, capped weight that is **gated/multiplied by
third-party engagement on that same repo** (stars/forks/external contributors), exactly
like Cadence only counts active days on validated repos. A solo unvalidated repo's craft
signals stay ≈0; the same signals on a validated repo unlock full weight. (Verified:
a 1-repo/0-PR account dropped from Craft p58 → p24; a 480★ repo with deep review stayed
high.) This is the same principle as §1 — self-controlled signals don't rank you.

**Influence / Reach** — Gold signal is **dependents** (others literally import the
code). **Implementation note:** the dependents graph ("Used by") is NOT in the official
API — either scrape it or, better, pull download stats from package registries (npm
registry API, pypistats, crates.io). Keep stars but quality-check them: stars *with*
forks + external contributors + issues from strangers = real; a spike of 5k stars with
0 forks / 0 issues over 2 days = star-buying (fetch stargazer timestamps via the
`Accept: application/vnd.github.star+json` header to inspect the curve shape).
Followers = low weight.

**Collaboration** — Where attribution shines: **merged PRs to repos the user does NOT
own**, especially popular ones (you can't unilaterally merge into someone else's repo →
a maintainer approved the work → high confidence). Plus reviews the user *gave* on
others' PRs, and cross-org breadth. Gaming guard: weight by *merged + substance +
reputation of the target repo* to defeat typo-PR / Hacktoberfest spam. One substantive
merged PR to a known repo >> 50 typo PRs.
> **KNOWN GAP (step 3, must close before going live — not a blocker now):** the
> external-PR filter compares `repository.owner` to the user's *personal* login. A user
> who owns/admins an **org** can self-merge a PR into that org's repo: `RepoOwner` is the
> org (≠ username), so it passes both filter layers and wrongly counts as Collaboration —
> exactly the non-attributable self-merge the system should exclude. Fix: also check the
> user's admin/ownership role on the target org/repo, or treat orgs the user controls as
> "own". Current tests cover the personal case only.

**Range** — NOT a count of languages, but **depth per language** (how much
externally-validated work exists in each). Render Specialist↔Generalist as a
*distribution shape* (concentration of validated work) — descriptive flavor for the
class, not "good/bad".

**Trust (meta-score, not a dimension)** — For each dimension, compute in parallel what
fraction of its underlying evidence is third-party-attributed vs self-controlled.
Trust = the weighted share of high-confidence evidence across the whole build. This is
the thing the user "levels up" by connecting GitHub. Honest, satisfying loop.
**The data model already supports this:** `evidence_items` carries
`verification_status` + `verification_confidence` (0–1), and `signal_events` carries a
`confidence_multiplier`. Confidence is already weighted at the per-evidence level — the
Trust meta-score aggregates exactly this. Reuse it; do not invent a parallel mechanism.

### Normalization & cold-start
Ranking is **percentile within the dev cohort** — but with ~10 users there's no
distribution. Pragmatic fix: seed a precomputed reference distribution from a sample of
public GitHub profiles, start with absolute thresholds, and migrate to true percentile
as the cohort grows. Make the normalization swappable behind an interface.

**LOCKED — reference population:** percentiles are measured against **active developers /
the target users who can demonstrate their level**, NOT all GitHub accounts. The median
*active* dev should land ~p50; the median GitHub account (mostly inactive, 1–2 repos, no
OSS) is NOT the baseline. Rationale: if the baseline is all of GitHub, every real user
clusters at p95+ and the score stops discriminating between the actual users — defeating
Explore and ranking. This decision drives all breakpoint calibration. (`reference.go`
should carry a comment stating this for whoever calibrates later.)

**Do NOT hand-tune breakpoints on 1–2 data points** — that's just a different guess and
over-fits noise. The seeded curve is a placeholder; real calibration happens when there's
enough cohort data to swap in the `CohortNormalizer`. Resist tuning the curve before then.

### GitHub API realities (save yourself pain)
- GraphQL v4 `contributionsCollection` gives commits/PRs/reviews/issues with
  **per-repository breakdown + the repository object**, so you can check each repo's
  owner & stars and do the "my repo vs someone else's repo" split the whole system
  depends on.
- OAuth **as the user** → use their token (their 5000/hr limit) and you can read private
  contribution counts if they grant it.
- Compute scores in a **batch/cron job and cache**; periodic recompute implements
  recency decay. Don't score synchronously on page load.
- (APIs drift — verify current rate limits / endpoint specs at implementation time.)

### Tuning backlog (revisit with real data — NOT decisions, not blockers)
- **Cap `repoWeight` in Collaboration.** Current `(log10(stars+1) − 0.95) / 1.5` has no
  ceiling; a PR to a mega-repo scales unboundedly. Consider `min(weight, ~2.0)` so one
  Linux-kernel PR doesn't swamp everything. Log already compresses it, so low priority.
- **Check Craft reference breakpoints.** Two different profiles (top OSS contributor and
  crafted solo dev) both saturate at p99 — the high end of the curve is flat and stops
  discriminating. Verify the upper breakpoints actually separate strong-from-stronger.

---

## 5. Aesthetic — high-fantasy MMORPG (WoW / GW2 / Lineage 2)

**Anchors:** World of Warcraft, Guild Wars 2, Lineage 2. NOT Diablo, NOT Destiny, NOT
DOTA — those skew dark/sci-fi/painterly and will pull the design the wrong way.

This is *ornate high fantasy*, a different design language from minimal dark themes:
- **Palette:** warm metallics (gold, bronze), parchment, deep blues / burgundy / forest
  green. Rich, not flat-black.
- **Panels:** ornate borders, decorative corner-pieces, framed stat panels (think the
  WoW character sheet's gilded frames; GW2's elegant art-nouveau edges).
- **Typography:** fantasy serif for headings/titles (NOT sci-fi mono). Readable body.
- **Stat bars:** "engraved" / inset feel, not flat progress bars.

**Icons:** use **game-icons.net** (https://game-icons.net) — thousands of CC-licensed
fantasy SVG icons (swords, shields, runes, classes, magic). SVG, so they compose
cleanly with the emblem generator. Give each class / dimension / stat its own icon.

**⚠️ Restraint warning:** ornate high-fantasy is *easier to make look cheap* than
minimal dark, if rushed — it degrades into "medieval clipart". Keep textures restrained
and the palette consistent. Quality over quantity of ornamentation. Read the repo's
`frontend-design` conventions if present.

---

## 6. Emblem generator

Deterministic, code-only (SVG), NO image API. Same build → same emblem (use a hash of
the build/scores as seed). The emblem is a procedurally-composed crest/sigil whose
shapes, colors, and motifs are functions of the 5 dimensions (e.g. a Collaboration-heavy
build reads visibly different from an Output-heavy one). This is more impressive to an
interviewer than calling an image API, scales to infinite users at zero cost, and needs
zero external assets beyond game-icons SVGs.

---

## 7. Pages (frontend)

1. **Landing** — what it is + a single "Connect GitHub" CTA. Where the aesthetic lives.
2. **Auth / OAuth callback** — GitHub login, fetch, a "forging your character…" loading
   state (good moment for the vibe).
3. **Character / Profile** — the flagship. Emblem, class/subclass/headline, the 5
   dimensions as framed stat bars, the Trust score, and the underlying
   artifacts/evidence. ~70% of design effort goes here.
4. **Explore** — grid of hunters, filter by class, sort. The "I'm not alone" moment.

---

## 8. Suggested implementation order

1. **Inventory** (§0) — report, then confirm. ✅ DONE.
2. **Scoring engine** with mock/fixture data — pure module, unit-tested, no external
   calls. Get the 5 dimensions + Trust producing sane numbers from hand-made fixtures
   first. ✅ DONE & VALIDATED — `backend/internal/scoring/` (pure, no DB/HTTP). Gaming
   guard, per-dimension decay (Decision 3), and non-dilutable Trust (Decision 4) all
   confirmed via sanity profiles incl. the durable-builder case (Output p10 / Influence
   p99 — decay is correctly per-dimension).
3. **GitHub OAuth + ingestion** — feed real data into the scoring engine. ✅ DONE.
   Engine kept pure (`Ingest` is the sole bridge to `GitHubInput`); `MockGitHubSource`
   via `MOCK_GITHUB`; GraphQL `contributionsCollection` (two dedup'd 365-day windows for
   the 548-day span), external-PR classification, reviews, star timestamps, CI/test
   detection, registry dependents. Scores computed but only logged (schema untouched).
   Scoring runs via a guarded, panic-recovering job with pollable status.
   *Carry-forward follow-ups (not blockers): (a) org self-merge gap in Collaboration —
   see §4; (b) scoring job status is in-memory, move to DB in step 4, and ensure the
   concurrency guard has a TTL so an orphaned "running" can't lock a user out.*
4. **AI generation layer** — scores → class/flavor (respect `MockAI`). Note: the schema
   migration to the new 5+Trust taxonomy (drop `trusted` column, rename dimensions,
   update `domain/signal` + `signal_service` mapping + AI prompt) lands as ONE change set
   around steps 3–4 per the Sequencing note in LOCKED DECISIONS.
   - **4a — schema migration + persistence:** ✅ DONE. New columns (raw + percentile per
     dimension, `trust`), old columns dropped, scores written by the scoring job (now
     DB-backed). Real-account checkpoint run: surfaced + fixed the Craft self-controlled
     gaming hole (validation gate, see §4); reference-population decision locked (see
     Normalization). Breakpoints intentionally NOT tuned (insufficient data).
   - **4b — AI rewiring:** ✅ DONE & VALIDATED. build_generator takes the 5+Trust
     percentiles (NOT raw) as primary input; CV is low-confidence fallback. Prompt uses
     an explicit p0–29/30–69/70–89/90–100 tier scale benchmarked against active devs, and
     an "all dimensions Low" rule forcing humble/emerging voice. Real-AI run confirmed
     discrimination: opposite shapes got different classes AND different voice
     (near-empty account → Pathfinder, "still taking shape", explicit Trust caveat;
     elite fixture → Architect, authoritative, no hedge). growth_paths reference real
     dimensions, confirming the model reads the shape. Permanent harness: `make harness`.
     Minor open item: confirm the AI model name is config-driven, not hardcoded.
5. **Frontend** — the four pages, wired to the backend, with the §5 aesthetic. Includes
   the full retheme from the current dark sci-fi palette to warm high-fantasy.
6. **Emblem generator** — last; it's polish on top of working scores.

Rationale: the scoring engine is the risky, valuable core — prove it in isolation before
coupling it to OAuth, AI, or UI.

---

## 9. The one rule to stay on scope

For every feature you're tempted to add, ask: **"Is this needed for GitHub → character →
Explore?"** If no, it goes in `VISION.md`, not the code.
