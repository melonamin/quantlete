# Project Review Remediation & Roadmap

## Overview

Remediation plan derived from the full project review (2026-07-19), which audited the Go backend, WASM mode + codegen, React frontend, tests/CI, and feature set vs. spec. The review found a strong architecture undermined by unenforced CI, confirmed data-loss and data-correctness bugs at the mode-parity seams, and bimodal test coverage. This plan was itself reviewed against the codebase (auto plan-review, 2026-07-19) and revised accordingly.

This plan is structured in four phases:

- **Phase 0 — Restore the safety net:** make CI actually run and green; clean the working tree.
- **Phase 1 — Data integrity:** fix confirmed data-loss and data-correctness bugs (WASM persistence, float truncation, exports, webhooks, SSE, training load, UTC timestamps).
- **Phase 2 — Hardening & coverage:** make generators fail loudly, harden the WASM bridge, close the highest-leverage test gaps, add frontend resilience, de-flake E2E.
- **Phase 3 — Product roadmap:** analytics trio, data portability, polish, docs, gamification. Each Phase 3 item gets its own detailed plan when picked up; this document tracks them at roadmap level only.

Phases must be completed in order. Within a phase, tasks are ordered by dependency; independent tasks may be reordered.

## Context (from discovery)

Findings come from the 2026-07-19 review; all file references below were re-verified by the plan review. Key references:

- CI: `.github/workflows/ci.yml` triggers on `main` but default branch is `master`; its go/e2e jobs lack the generator steps (`*.gen.go` is gitignored → fresh checkout cannot compile); `build.yml` has generators but zero test steps; `release.yml:24` pins Go 1.23 vs `go.mod` 1.25.4; Deploy Demo workflow (`demo.yml`, weekly cron) failing since 2026-01-25.
- WASM persistence: Go writes bypass the `dirty` flag in `web/src/lib/wasm/db/sql-js.ts:121-127`; nothing calls `persistDatabase()` after imports/mutations (`go-provider.ts` has it at :1892, unused by `importCompleteCallback`); no `pagehide`/`visibilitychange` flush; no `navigator.locks` anywhere.
- WASM driver: `go-sqlite3-js` truncates all REAL values to int (`dest[i] = jsVal.Int()` for `js.TypeNumber`, `sqlite3.go:~197`); no fork/replace in `go.mod`.
- Backend: export handler (`internal/api/handlers/export.go`) missing athlete scoping, `MaxExportActivities=50000` clamped to 200 by `Normalize()`, wired straight to the repo (service-layer bypass, `router.go:240`); webhook handler (`webhooks.go:106-107`) cancels async imports via `defer cancel()` on a 5-min ctx and skips subscription-ID validation when `WebhookSubscriptionID == 0` (`webhooks.go:88`); SSE killed by `WriteTimeout` 15s (`serve.go:253`, `config.go:55`) and `middleware.Timeout(60s)` (`router.go:196`); training-load EWMA seeded from window start with swallowed upsert error (`training_load.go:384`); text timestamps stored with local offsets (`helpers.go:84,93`); rate limiter over-waits flat 15m/24h (`ratelimit.go:117,120`; `TimeUntil15MinReset()` exists at :223).
- Generators: `scripts/generate-adapters/main.go` silently skips `*float64` query params (`MinDistanceM`/`MaxDistanceM` absent from `adapters.gen.go`) and skips bool defaults on the WASM side only (`main.go:516`).
- Frontend: no error boundary / 404 route; 20 statically-imported pages in `router.tsx`; silent mutation failures in settings/FTP/weight; `web/src/lib/api/notifications.ts` bypasses the envelope (raw fetch, :15,:28); server SSE client drops `waiting_*` fields (`sse-client.ts:40-57`); calendar buckets on `start_date` not `start_date_local` (`month-view.tsx:44`); heatmap country selector no-op; pace `:60` rounding in `lib/format.ts`.
- Coverage: `internal/services` 15.6%, `internal/storage` 11.8%, `internal/importer` 8.6%, zero for `internal/api` middleware, `internal/strava`, `internal/pagination`, `cmd/wasm`; 2 web unit test files; E2E has exactly 127 `waitForTimeout` calls (worst: `journeys/content-browsing.spec.ts` 13, `error-states.spec.ts` 12, `pages/{segments,photos,activities}.page.ts` 7 each), server mode only.

## Development Approach

- **testing approach**: TDD when possible (per project guidelines); at minimum every task ships tests with the code
- complete each task fully before moving to the next
- make small, focused changes; one commit per task (or smaller)
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - tests are not optional — they are a required part of the checklist
  - unit tests for new and modified functions; cover success and error scenarios
  - use the existing no-mocks harness: in-memory SQLite with full migrations (`testDB()` pattern in `internal/storage/storage_test.go`)
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- run `just generate` after any service/schema change; run `just check-adapter-parity` after adapter annotation edits; `schema.gen.ts` and other `*.gen.ts` files are committed — regenerate and commit them with the change
- **Raw SQL policy for this plan:** Tasks 10 and 18 modify *existing* hand-written SQL in `internal/storage` (training-load inline query, activities LIKE clause) in place — treated as maintenance of grandfathered queries. Any *new* queries go through `schema/queries` codegen. ✅ Approved by Sasha 2026-07-19 (in-place edits to grandfathered queries).
- maintain backward compatibility of the API envelope and DB schema (migrations only, never edit applied migrations)

## Testing Strategy

- **unit tests**: required for every task (see Development Approach)
- **e2e tests**: project has Playwright E2E (`web/tests/e2e/`, run via `just test-e2e`)
  - UI changes → add/update page objects + specs in the same task
  - backend changes visible in UI (SSE, exports, sync) → add/update E2E in the same task
  - Task 4 adds a WASM-mode E2E project — after it exists, WASM-visible changes must cover it too
- full gates: `just test` (Go + web unit), `just lint`, `cd web && yarn typecheck`, `just test-e2e`

## Progress Tracking

- mark completed items with `[x]` immediately when done
- add newly discovered tasks with ➕ prefix
- document issues/blockers with ⚠️ prefix
- update plan if implementation deviates from original scope
- keep plan in sync with actual work done

## What Goes Where

- **Implementation Steps** (`[]` checkboxes): tasks achievable in this codebase — code, tests, docs
- **Post-Completion** (no checkboxes): external actions — Cloudflare/GitHub settings, manual verification, deployed-site checks

---

## Implementation Steps

## Phase 0 — Restore the safety net

### Task 1: Make CI run on master and pass

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`
- Modify: `justfile`

- [x] change `ci.yml` triggers from `main` to `master` (push + pull_request)
- [x] add code-generation steps before any `go build`/`go test` in the `go` and `e2e` jobs (mirror the generator sequence in `build.yml`: `generate-sql`, `generate-adapters`, `generate-wasm-registration`), since `*.gen.go` is gitignored and a fresh checkout does not compile
- [x] remove `CGO_ENABLED=1` from `ci.yml` (project is pure Go / `modernc.org/sqlite`)
- [x] align `release.yml` Go version with `go.mod` (1.25.x) — update the runner, do not downgrade the project
- [x] add `yarn test:unit` (vitest) to the web job
- [x] widen the justfile `test-go`/`test-go-cover`/`lint-go` recipes from `./cmd/... ./internal/...` to `./...` so `scripts/*` (home of Task 12's generator tests) is covered locally; ci.yml already runs `./...`
- [x] extract the E2E setup shared by CI and `just test-e2e` so the two harnesses stop drifting (CI runs Playwright natively, local runs via `docker-playwright.sh` — extraction, not forcing CI into Docker; shared script: `scripts/e2e-setup.sh`)
- [x] ➕ stub `web/dist` in the ci.yml `go` job before build/test — root `embed.go` (`go:embed web/dist/*`) fails to compile on a fresh checkout without it (same failure class as the Deploy Demo breakage)
- [x] ➕ fix 43 pre-existing lint issues in `scripts/*` surfaced by the `./...` widening (errcheck, gosec, gocritic, staticcheck, gofmt, unused, ineffassign) — CI lint job fails without this; 6 were in `internal/` (OAuth redirect leak, cookie Secure, multipart cap, path containment, importer ctx annotation) and shipped with tests
- [x] ➕ first-run CI fixes: pin golangci-lint-action@v8 + linter v2.10.1 (v6/latest installs a Go 1.24-built v1.x that rejects go.mod 1.25.4); prettier --write 14 web files that never had `prettier --check` enforced (includes committed `*.gen.ts` — generators emitting non-prettier output is a latent issue for Task 12)
- [x] ➕ fix production bug caught by CI's first E2E run: `SUM(calories)` (REAL) scanned into generated `int` — calendar range/summary + dashboard summary 500 on fractional totals. Fixed via `CAST(... AS INTEGER)` in `schema/queries/{calendar,dashboard}.sql` + regeneration + fractional-calories regression test. This is the Task 12 generator-inference flaw manifesting; Task 12 must make inference schema-aware
- [x] push a branch and verify every ci.yml job goes green before merging — ✅ 2026-07-20, run 29750087786: Go, Go WASM, Web, E2E (8m30s), WASM E2E (2m03s), Build Web, Build Binaries all pass on PR #24
- [x] verify `just test`, `just lint` pass locally with the same package scope CI uses (Go tests PASS, vitest 38/38 PASS, lint pending the ➕ scripts fixes)

### Task 2: Diagnose and fix the Deploy Demo workflow

**Files:**
- Modify: `.github/workflows/demo.yml` (and/or `deploy.yml` — determine from failure logs)

- [x] pull logs for the last failed runs and identify the root cause — "Generate demo database" compiles the root package via `go run ./cmd/quantlete`, but `embed.go` (`go:embed web/dist/*`) fails on a fresh checkout; every run since 2026-01-25 died there in ~40s. Bonus finding: GitHub auto-disabled the weekly cron after 60 days of repo inactivity (no runs since 2026-03-29)
- [x] fix the workflow (or the deploy target config) accordingly — stub `web/dist` before the demo-DB step (deployed site builds to `dist-demo`, embed assets unused)
- [ ] trigger a manual run and verify it succeeds end to end (after merge to master; also re-enable the auto-disabled schedule: `gh workflow enable demo.yml`)
- [ ] verify demo.quantlete.fit serves the freshly deployed demo (post-merge)
- [x] add a failure notification (e.g. workflow failure → GitHub issue or email) so silent weekly failures cannot recur — `if: failure()` step comments on/creates a "Deploy Demo workflow failed" issue with minimal `issues: write` permission

### Task 3: Working-tree hygiene and TODO refresh

Most root artifacts are already gitignored/untracked — this is local cleanup plus ignore hardening; only two files are actually git-tracked and need removal.

**Files:**
- Delete (untracked, local): `coverage.out`, `coverage.html`, `progress-2026-01-21-github-issues-roadmap.txt`, `progress-plan-add-blue-bird-gif-1769547895747071262.txt`, `wasm` (12.4MB root binary), `docs/plans/2026-01-26-session-cancelled.md`
- Untrack + delete (git-tracked): `internal/storage/queries.gen.go.unformatted`, `cmd/wasm/rewind_extract.txt`
- Modify: `.gitignore`, `TODO.md`, `justfile`

- [x] delete the untracked stale artifacts and the cancelled-session plan
- [x] `git rm internal/storage/queries.gen.go.unformatted cmd/wasm/rewind_extract.txt`; add an ignore pattern for `*.unformatted`
- [x] point `just test-go-cover` output at `tmp/` (gitignored) instead of repo root (+ `mkdir -p tmp` for fresh clones)
- [x] update `TODO.md` to reality: 13.3 merged (PR #8), 13.1 backend AND most frontend done (verified against web/src — only `GearROISummary` remains), Documentation Milestones checked off where docs exist
- [x] verify `just build` still succeeds (build-go verified)
- [x] run full test suite — must pass before Phase 1 (Go ./... PASS, vitest PASS)

## Phase 1 — Data integrity

### Task 4: WASM-mode E2E harness (with red persistence spec)

Built first so Task 5's fix can be verified red → green.

**Files:**
- Create: `web/tests/e2e/wasm/` (specs + fixtures)
- Modify: `web/playwright.config.ts`, `justfile`, `.github/workflows/ci.yml`

- [x] add a Playwright project that serves the WASM build (demo DB seed) — Chromium only (OPFS support); keyed on `WASM_E2E=1` env (argv sniffing breaks in Playwright worker processes)
- [x] smoke specs: app boots, dashboard renders, activities list renders
- [x] persistence spec: dashboard widget toggle → reload → assert survival — complete test, `test.fixme()` with TODO(Task 5)
- [x] float-value spec: non-integral value assertion — `test.fixme()` with TODO(Task 6)
- [x] wire into `justfile` (`test-e2e-wasm`) and add to ci.yml (`wasm-e2e` job)
- [x] run the smoke specs — 2 passed, 2 fixme-skipped as designed
- [x] ➕ repair the stale server-mode E2E suite (found when CI first ran it): 93/266 failing from UI drift since January — `[class*="card"]` ambiguous vs shadcn data-slots (63 uses), badges navigating the backend `/badges` route, badge/filter class-token false positives, power dropdown left open, demo-mode import controls. Two Codex rounds + DOM-snapshot evidence → 266/266 pass, 0 flaky locally
- [x] ➕ make E2E server port configurable (`E2E_PORT`, default 8081) so local runs don't collide with a developer's live instance

### Task 5: WASM persistence — Go writes must reach OPFS

**Files:**
- Modify: `web/src/lib/wasm/go-storage.ts`
- Modify: `web/src/lib/wasm/db/sql-js.ts`
- Modify: `web/src/lib/data/wasm/go-provider.ts`
- Create: `web/src/lib/wasm/db/sql-js.test.ts` (or extend existing)

- [x] mark the database dirty from the Go write path — Proxy bridge wrapping the sql.js Database handed to go-sqlite3-js: `Statement.run` and (conservatively) `Database.exec` mark dirty
- [x] call `persistDatabase()` in `importCompleteCallback` (incl. failed imports with partial data) and after every mutating `GoWasmProvider` method via a `persistAfter` wrapper
- [x] add a `pagehide`/`visibilitychange` flush (registered at init, cleaned up on provider dispose — disposal wired through the React data-provider lifecycle)
- [x] take a Web Lock around init and persist (feature-detected; lock-only scope; second-tab UX still deferred). Persist clears `dirty` before the async save and restores it on failure so in-flight writes aren't lost
- [x] write unit tests: 43 vitest cases across bridge dirtying, mutations, import completion, pagehide/disposal, lock scope
- [x] remove the TODO from Task 4's persistence spec and verify it now passes — ➕ required splitting the harness: the demo build intentionally never persists, so a second `wasm-persist` Playwright project now serves the REAL WASM build (dist-wasm, port 4175, empty OPFS); the spec uses the settings unit toggle (only Go-write reachable pre-import, verified by live probe) → toggle, reload, survives. Passes; demonstrably failed pre-fix
- [x] run tests — vitest 43 PASS, WASM projects 3 pass/1 fixme-skip, server suite re-run 266/266 (provider lifecycle shared with server mode)

### Task 6: Fix go-sqlite3-js float truncation

**Files:**
- Create: fork/vendored patch of `github.com/matrix-org/go-sqlite3-js`
- Modify: `go.mod`, `go.sum`

- [x] reproduce first: confirmed via Task 4's red float spec (demo seed speed 8.28298918... rendered truncated pre-fix)
- [x] decide fork hosting: in-repo vendored `third_party/go-sqlite3-js/` + relative `replace` — reproducible CI, no external repo to maintain; rationale + upstream commit recorded in its README
- [x] patch the scan path: `js.TypeNumber` → float64 when fractional, int64 otherwise; single commented change against frozen upstream
- [x] add the `replace` directive in `go.mod` (+ `go mod tidy` — vendored module's test deps entered go.sum)
- [x] audit blast radius: 52 REAL columns across 13 tables enumerated; all generated and hand-written scan destinations float-capable; the four int calorie aggregates already CAST — no further fixes needed
- [x] verification: float spec fixme removed and PASSES against the rebuilt WASM binary; full WASM suite 4/4 green (no fixmes remain)
- [x] run tests — go build (native + js/wasm), vet, lint 0 issues (third_party excluded as nested module), internal tests pass

### Task 7: Fix export endpoints (athlete scoping + truncation)

**Files:**
- Modify: `internal/api/handlers/export.go`
- Modify: `internal/services/` (route export through the service layer)
- Modify: `internal/pagination/pagination.go`
- Create/Modify: export handler/service tests

- [x] route CSV/JSON export through `ActivityService` — new `StreamExport(ctx, in, writePage)`: requires non-zero AthleteID, fetches Normalize()-bounded 200-row pages, hands each page to the encoder before fetching the next (bounded memory)
- [x] replace the single clamped query with the pagination loop; deleted `MaxExportActivities`
- [x] removed dead `ExportHandler.ExportStats`; `ParseOrderDir` now case-insensitive
- [x] write tests: athlete isolation (two athletes seeded), 250-row completeness (clamp regression), CSV+JSON paths, CSV header, missing athlete ID, order-dir casing — real-SQLite harness
- [x] E2E: shared 50-activity seed left unchanged deliberately (raising it slows every suite run); the clamp regression is covered at the service/handler level — noted as the accepted trade-off vs. the original plan line
- [x] run tests — services/handlers/pagination pass, lint 0 issues

### Task 8: Fix webhook processing (context lifetime + forged events)

**Files:**
- Modify: `internal/api/handlers/webhooks.go`
- Modify: `internal/api/handlers/webhooks_test.go`

- [x] for `create` events, pass `context.Background()` to the async `Importer.Start`; `update`/`delete` keep the bounded 5-min ctx; other `Importer.Start` call sites audited (already correct)
- [x] refuse events when `WebhookSubscriptionID` is unconfigured: warn + 200-ack + drop (an error response would make Strava disable the subscription); mismatch still 403
- [x] ack first, then non-blocking semaphore admission: only admitted events spawn goroutines; saturated events warn + drop (scheduled pull sync is the catch-up path) — no goroutine pileup, ack never delayed
- [x] write tests: import context outlives handler; unconfigured-subscription drop; mismatch rejection; verified-delete; prompt 2xx under saturation — all race-enabled
- [x] run tests — pass with -race, lint 0 issues

### Task 9: Fix SSE lifetime and rate-limit progress parity

**Files:**
- Modify: `cmd/quantlete/serve.go`, `internal/api/router.go`
- Modify: `web/src/lib/data/server/sse-client.ts`
- Modify: `web/src/lib/data/server/sse-client.test.ts`

- [ ] exempt the SSE route (`/api/v1/import/events`, `router.go:320`) and export downloads from `middleware.Timeout(60s)`; handle the write deadline per-route (`http.ResponseController.SetWriteDeadline` bumped on every write/keepalive) instead of relying on the global 15s `WriteTimeout`
- [ ] verify a long-lived SSE connection survives >60s with keepalives (slow import or keepalive-interval test hook)
- [ ] map `waiting_for_rate_limit`, `waiting_until`, `waiting_reason` in `sse-client.ts:40-57` so `RateLimitCountdown` works in server mode (WASM already maps them)
- [ ] write/extend tests: Go test for timeout exemption wiring; vitest for the SSE field mapping (extend `sse-client.test.ts`, and fix its max-backoff test at :409 to actually assert the cap)
- [ ] run tests — must pass before next task

### Task 10: Fix training-load EWMA seeding

⚠️ Touches existing hand-written SQL in `training_load.go` — see the raw SQL policy note in Development Approach (needs Sasha's confirmation).

**Files:**
- Modify: `internal/storage/training_load.go`
- Modify: `internal/services/stats.go` (if the window plumbing changes)
- Modify: `internal/storage/training_load_test.go`

- [ ] seed CTL/ATL from the athlete's full history (compute from first activity, or resume from the stored day preceding the requested window) so values are range-independent
- [ ] make `daily_training_load` persistence range-independent; make `upsertDaily` transactional and stop swallowing its error (`training_load.go:384`)
- [ ] write tests: same-day CTL/ATL identical across different requested ranges; early-window values match full-history computation; persisted summary stable after overlapping range requests
- [ ] run tests — must pass before next task

### Task 11: Store and compare all timestamps in UTC

⚠️ Highest data-risk task in the plan: the migration auto-runs in browsers against users' OPFS databases (`cmd/wasm/main.go:77` migrates on every load) with no backup path. Task 6 must be complete first — in WASM this migration executes through the forked driver's scan path.

**Files:**
- Modify: `internal/storage/helpers.go`
- Create: `schema/migrations/007_utc_timestamps.sql`
- Modify: `internal/storage/helpers_test.go`
- Regenerate: `web/src/lib/wasm/db/schema.gen.ts` (committed)

- [ ] normalize to UTC inside `TimeToSQL`/`Value` (`t.UTC().Format(...)`, `helpers.go:84,93`) so every stored text timestamp carries `+00:00` (`SQLiteTimePtr`/`SQLiteTimeToSQL` delegate here, so this covers them)
- [ ] audit filter construction (`AddTimeRangeFilter`, `internal/shared/parse.go`) to normalize parsed inputs to UTC before formatting
- [ ] write migration `007_utc_timestamps.sql` normalizing existing offset-bearing `created_at`/`updated_at`-style values to UTC — **idempotent and transactional**, with row-count/spot-value assertions in its test; verify against a copy of a real server DB first
- [ ] run `just generate` and commit the regenerated `schema.gen.ts` so the browser schema includes the new migration (skipping this silently diverges the modes)
- [ ] test the migration under sql.js/WASM (via the Task 4 harness: load a pre-007 fixture DB, boot, assert values migrated correctly) — this is the OPFS-safety check
- [ ] document the rollback story (007 is value-normalizing and idempotent; re-running is safe; restoring = re-import) before shipping
- [ ] write tests: round-trip through `Value`/scan preserves the instant; range filter with a `+05:00` input matches the correct UTC rows; ordering by `created_at` correct across mixed historical offsets
- [ ] run full test suite (including full WASM E2E) — must pass before Phase 2

## Phase 2 — Hardening & coverage

### Task 12: Make code generators fail loudly

**Files:**
- Modify: `scripts/generate-adapters/main.go`
- Modify: `scripts/generate-sql/main.go`
- Create: generator tests under `scripts/`

- [ ] `generate-adapters`: unsupported query-param field types (e.g. `*float64`) are a build error, not a silent skip; add `float64`/`*float64` support so `min_distance_m`/`max_distance_m` work over HTTP
- [ ] `generate-adapters`: reject `bool ... default=` (HTTP applies it, WASM skips it at `main.go:516` — latent divergence)
- [ ] regenerate (`just generate`) and verify distance filters now appear in `adapters.gen.go`; add a unit/E2E assertion that server-mode distance filtering works
- [ ] `generate-sql`: validate inferred column types against the actual schema (or at minimum warn loudly on heuristic-only inference)
- [ ] write generator tests covering the new failure modes (unsupported type → error; bool default → error) — runnable via `just test-go` thanks to Task 1's scope widening
- [ ] run `just generate && just check-adapter-parity` and full test suite — must pass before next task

### Task 13: WASM runtime panic recovery and error surfacing

**Files:**
- Modify: `cmd/wasm/import.go`, `cmd/wasm/helpers.go`

- [ ] add recovery to the bare background goroutines in `cmd/wasm/import.go` (event forwarder at :104, import goroutine at :118) so a panic cannot kill the WASM runtime mid-import
- [ ] make `recoverPanic` surface the panic message via named return (`errorJSON`) instead of returning undefined to JS
- [ ] write tests: panicking import path recovers and reports an error state (Go unit where feasible); verify WASM E2E suite still green
- [ ] run tests — must pass before next task

### Task 14: WASM namespace and rate-limit info plumbing

**Files:**
- Modify: `web/src/lib/wasm/go-storage.ts`, `web/src/lib/wasm/strava/client.ts`
- Modify: `cmd/wasm/import_strava_adapter.go`, `cmd/wasm/import.go`

- [ ] add `getRateLimitInfo`, `onImportProgress`, `onImportComplete` to the frozen `__quantlete_go_wasm__` namespace — they must be present when the namespace is **constructed** (before `loadGoWasm()`); `Object.freeze` prevents adding them later
- [ ] read them from the namespace on the Go side (`import_strava_adapter.go:238`, `import.go:256,273`) — fixes always-zero rate-limit info/ETA in WASM imports
- [ ] write tests: Go-side plumbing unit test; WASM E2E assertion that import progress carries non-zero rate-limit info
- [ ] run tests — must pass before next task

### Task 15: WASM init readiness and event fan-out

**Files:**
- Modify: `web/src/lib/wasm/go-storage.ts`, `cmd/wasm/main.go`
- Modify: `web/src/lib/data/wasm/go-provider.ts`

- [ ] replace the 100ms init sleep (`go-storage.ts:~200`) with an explicit ready signal resolved from Go's `main()`
- [ ] back `GoWasmProvider.subscribeToEvents` with the existing `DataEventEmitter` (multi-listener safe; either unsubscribe currently clears both); stop swallowing JSON parse errors
- [ ] write tests: multi-subscriber emitter behavior (vitest); init resolves without arbitrary sleep (harness-level)
- [ ] run tests — must pass before next task

### Task 16: Calendar summary parity (WASM)

**Files:**
- Modify: `internal/services/dashboard.go` (or wherever the summary lives — add `//wasm:export`/adapter as needed)
- Modify: `web/src/lib/wasm/go-storage.ts` (delete the TS reimplementation at :~581-624)
- Regenerate: `cmd/wasm/registration.gen.go`, goStorage TS wrappers (committed)

- [ ] route `getCalendarSummary`/`getCalendarActivities` through the generated Go function instead of the hand-written TS reimplementation with hardcoded zeros (`total_calories: 0, workout_count: 0, challenges_completed: 0`)
- [ ] add the required `//wasm:export`/adapter annotation, then `just generate-wasm-registration && just generate-go-storage` (and `just check-adapter-parity`); commit regenerated files
- [ ] write tests: WASM E2E asserts calendar summary shows real (non-zero) values against the demo seed; Go unit for the service method if new
- [ ] run tests — must pass before next task

### Task 17: Backend tests — security surface + rate limiter fix

**Files:**
- Create: `internal/api/middleware_test.go` (CSRF, security headers)
- Create: `internal/pagination/pagination_test.go`
- Create: `internal/strava/ratelimit_test.go`
- Modify: `internal/strava/ratelimit.go`

- [ ] tests for CSRF middleware: Origin==Host allowed, cross-origin blocked, Referer fallback, GET untouched
- [ ] tests for security headers middleware (CSP present)
- [ ] tests for `pagination.Normalize()` bounds and case-insensitive `ParseOrderDir` (from Task 7)
- [ ] fix rate limiter waits while writing its tests: use `TimeUntil15MinReset()` (`ratelimit.go:223`) at the 15-min limit and midnight-UTC reset for the daily limit instead of flat 15m/24h sleeps (`ratelimit.go:117,120`); tests cover window-aligned wait durations
- [ ] tests for the Strava client 401-refresh path and `UpdateCredentials` race (run with `-race`)
- [ ] run tests — must pass before next task

### Task 18: Backend behavior fixes — LIKE escaping and SaveStream cap

⚠️ Touches existing hand-written SQL in `activities.go` — see the raw SQL policy note (needs Sasha's confirmation).

**Files:**
- Modify: `internal/storage/activities.go`
- Modify: `internal/services/activities.go`
- Create/Modify: tests alongside each

- [ ] escape LIKE wildcards (`%`, `_`) in activity search (`activities.go:273-275`) with tests proving literal-match behavior
- [ ] enforce the decoded-length cap in `SaveStream`: validate actual `len(in.Data)` against `maxStreamDataSize`, not the client-asserted `original_size`; tests for at-cap, over-cap, and spoofed-size cases
- [ ] run tests — must pass before next task

### Task 19: Backend coverage sweep — importer, storage, services

**Files:**
- Modify/Create: tests under `internal/importer/`, `internal/storage/`, `internal/services/`

- [ ] importer: phase transitions, resume/watermark, per-phase counters, cancellation, completion event payload (extend the existing importer test using the real-SQLite harness)
- [ ] storage: repository CRUD tests for the untested high-traffic repos (activities filters/sort/search, streams, segments, athlete metrics) using `testDB()`
- [ ] services: tests for activities, stats, dashboard happy-path + authz (`ErrForbidden` on foreign `AthleteID`)
- [ ] target: `internal/services` and `internal/storage` above 50% each; record actual numbers here when done
- [ ] run full test suite — must pass before next task

### Task 20: Frontend resilience — error containment and code splitting

**Files:**
- Create: `web/src/components/layout/error-boundary.tsx`
- Modify: `web/src/routes/router.tsx`, `web/src/main.tsx`
- Modify: `web/src/components/dashboard/widget-grid.tsx` (per-widget boundary)
- Create: tests for the boundary + a route smoke test

- [ ] add an app-level ErrorBoundary and a router `defaultNotFoundComponent` / `defaultErrorComponent` (also unblocks the pending animated-cat-404 plan's mount point)
- [ ] wrap each dashboard widget in a boundary so one throwing widget degrades to an inline error card instead of blanking the app
- [ ] convert the 20 statically-imported pages in `router.tsx` to lazy route components; verify ECharts and Leaflet leave the initial chunk (check `yarn build` output / bundle visualizer)
- [ ] memoize chart `option` objects, starting with `activity-stream-profile.tsx` and `elevation-profile.tsx`; replace `Math.min(...spread)`/`Math.max(...spread)` over stream arrays with a loop
- [ ] write tests: boundary renders fallback on child throw; 404 route renders for unknown URL
- [ ] run vitest + typecheck + E2E smoke — must pass before next task

### Task 21: Frontend correctness batch — mutations, parity seams, small bugs

**Files:**
- Modify: `web/src/lib/api/notifications.ts`, `web/src/components/settings/ftp-editor.tsx`, `web/src/components/settings/weight-editor.tsx`, `web/src/pages/settings.tsx`, `web/src/pages/athlete.tsx`
- Modify: `web/src/components/calendar/month-view.tsx`, `web/src/pages/heatmap.tsx`, `web/src/components/dashboard/widget-grid.tsx`
- Modify: `web/src/lib/format.ts`

- [ ] route `notifications.ts` through `ApiClient` (envelope unwrap, error handling, credentials) so the notification test indicator actually works in both modes
- [ ] standardize mutation error UX: shared toast/`onError` for settings editors and widget-grid autosave (with layout rollback); stop clearing FTP/weight inputs before the mutation settles (follow the `strava-credentials-form.tsx` pattern)
- [ ] bucket calendar chips on `start_date_local` (already selected by the query) so day cells, month list, and chips agree
- [ ] fix the heatmap country selector to actually filter bounds by the selected country (`handleCountrySelect` currently identical in both branches)
- [ ] fix pace rounding in `lib/format.ts` (`4:60/km` bug — round total seconds first)
- [ ] write tests for each fix (vitest for format/notifications envelope; E2E for settings save error feedback and calendar bucketing where practical)
- [ ] run tests — must pass before next task

### Task 22: Frontend unit-test foundation

**Files:**
- Create: `web/src/lib/format.test.ts`, `web/src/lib/maps/polyline.test.ts`, store tests under `web/src/stores/`
- Modify: `web/vite.config.ts`, `web/src/lib/data/events.test.ts`

- [ ] tests for all `lib/format.ts` functions (would have caught the `:60` bug) — success + edge cases (zero, negative, huge, unit systems)
- [ ] tests for `lib/maps/polyline.ts` decode
- [ ] tests for the Zustand stores (filters, dashboard, onboarding persistence behavior)
- [ ] widen the vitest `include` glob to `src/**/*.test.{ts,tsx}`
- [ ] replace the tautological constants-sync block in `events.test.ts:262-279` with a real cross-language check (or delete it)
- [ ] run vitest — must pass before next task

### Task 23: De-flake the E2E suite

**Files:**
- Modify: `web/tests/e2e/journeys/content-browsing.spec.ts` (13 hard waits), `web/tests/e2e/error-states.spec.ts` (12), `web/tests/e2e/pages/segments.page.ts`, `pages/photos.page.ts`, `pages/activities.page.ts` (7 each), `web/tests/e2e/fixtures/index.ts`, remaining specs/pages with hard waits

- [ ] replace `waitForTimeout` calls (127 total) with auto-waiting assertions (`expect(...).toBeVisible()` / `toHaveText()`), starting with the worst offenders above
- [ ] convert `if (await x.isVisible())` conditional assertions into deterministic assertions backed by known demo-seed data
- [ ] remove the blanket 100ms wait in `fixtures/index.ts:56`
- [ ] run the full suite 3x locally (`PLAYWRIGHT_WORKERS=1` and default) — zero flakes tolerated
- [ ] run in CI and confirm green — must pass before Phase 3

### Task 24: Verify acceptance criteria (Phases 0–2)

- [ ] all Phase 0–2 checkboxes complete
- [ ] CI green on master across all jobs (go, web, e2e, wasm-e2e)
- [ ] WASM persistence + float-parity E2E specs passing (no remaining TODO-marked red specs)
- [ ] re-run the review's confirmed repro cases: >200-activity export complete; webhook create event triggers a real sync; SSE survives >60s; CTL/ATL range-independent; calendar summary non-zero in WASM
- [ ] coverage targets met (`internal/services`/`internal/storage` ≥50%; web unit tests cover format/polyline/stores)
- [ ] full gates: `just test`, `just lint`, `cd web && yarn typecheck`, `just test-e2e`, `just test-e2e-wasm`

### Task 25: Update documentation and close out

**Files:**
- Modify: `ARCHITECTURE.md`, `CLAUDE.md`, `docs/_sidebar.md`
- Create: `docs/getting-started/browser-mode.md` (or similar), `docs/deployment/macos.md`

- [ ] fix ARCHITECTURE.md drift: pure-Go/no-CGO, port 8081, remove the phantom `export` CLI command, current troubleshooting layout
- [ ] fix CLAUDE.md false claims: WASM runs on the main thread (not a web worker); `generate-sql` does not emit `queries.gen.ts`
- [ ] write the "Browser-only mode" docs chapter (OPFS data location + caveats, worker OAuth, multi-tab behavior after Task 5, persistence guarantees)
- [ ] document the macOS menu-bar app as a distribution channel
- [ ] update this plan's checkboxes and move it to `docs/plans/completed/`

## Phase 3 — Product roadmap (each item gets its own plan when picked up)

### Task 26: Analytics trio (TODO 13.4–13.6)

- [ ] create detailed plan: aerobic decoupling (`algorithms/go`, `activity_analytics` table, `DecouplingBadge`)
- [ ] create detailed plan: race time predictor (Riegel; `/race-predictions` page + widget)
- [ ] create detailed plan: efficiency index over time
- [ ] implement per those plans (each with unit + E2E tests)

### Task 27: Data portability

- [ ] create detailed plan covering: per-activity GPX export, full-DB backup/restore UX, and WASM(OPFS)↔server migration (shared schema makes this near-free; it also de-risks single-browser OPFS storage)
- [ ] implement per that plan

### Task 28: Polish (TODO 12.10) before new UI

- [ ] create detailed plan: mobile responsiveness audit, accessibility (ARIA/keyboard) audit, Lighthouse/bundle performance audit (Task 20's code splitting is the head start)
- [ ] implement per that plan
- [ ] then (and only then) pick up the pending cosmetic plans (404 cat, empty-state spaceship, welcome bird, cab loader) — the 404 page slots into Task 20's `defaultNotFoundComponent`

### Task 29: Gamification release

- [ ] create one combined plan folding `docs/plans/2026-01-27-custom-achievements.md` into Explorer Tiles (TODO Phase 14) so gamification ships as a coherent release
- [ ] implement per that plan

## Technical Details

- **WASM persistence flow (target state):** Go write → bridge marks DB dirty → 30s timer OR mutation/import-complete hook → `persist()` under Web Lock → OPFS. `pagehide` flush as backstop. Second-tab read-only UX deferred to a follow-up task.
- **go-sqlite3-js patch:** scan path returns `int64` when `v == math.Trunc(v)` else `float64`; keep the fork minimal (one commit) for easy upstream rebases; hosting decision (hosted fork vs vendored + relative `replace`) recorded in Task 6. Blast radius: int-typed scan destinations over REAL columns must be audited — exact-only float64→int64 conversion.
- **Export streaming:** iterate pages of `Normalize()`-bounded size inside the service, writing rows to the CSV/JSON encoder per page — never a single unbounded query, never a silent clamp.
- **UTC migration:** only `created_at`/`updated_at`-style columns written via `TimeToSQL(time.Now())` carry local offsets; `start_date` from Strava is already UTC. Migration 007 rewrites offset-bearing values to their UTC equivalent — idempotent, transactional, auto-runs in browsers (`cmd/wasm/main.go:77`), so it must be tested under sql.js with a pre-007 fixture before shipping; `schema.gen.ts` must be regenerated and committed alongside.
- **SSE deadlines:** per-connection `SetWriteDeadline` bumped on every write/keepalive; router-level timeout middleware wraps everything *except* the SSE and export routes.
- **Generator policy:** unknown/unsupported constructs are errors. Silence is the failure mode that caused the distance-filter and bool-default divergences — never skip.
- **Frozen namespace:** `__quantlete_go_wasm__` is frozen at construction; all Go-visible callbacks (`stravaFetch`, `getRateLimitInfo`, `onImportProgress`, `onImportComplete`) must be defined before `loadGoWasm()`.

## Post-Completion

**Manual verification:**
- Real Strava import end-to-end in WASM mode on a throwaway account: multi-hour import, close tab mid-import, reopen — verify resume + no data loss
- Webhook flow against a real Strava push subscription with `WebhookSubscriptionID` configured
- Verify demo.quantlete.fit, quantlete.fit (landing), and docs.quantlete.fit after Deploy Demo fix
- Load-test SSE with an artificially slow import to confirm no reconnect loops
- Spot-check a real user-scale OPFS database migrates cleanly through 007 (browser devtools → OPFS inspection)

**External system updates:**
- If the GitHub default branch is ever renamed to `main`, revisit workflow triggers (or use a branch-agnostic trigger)
- Cloudflare Worker (`worker/`) unaffected by this plan, but re-verify OAuth proxy after any WASM auth-flow changes in Tasks 14/15
- Consider enabling branch protection on `master` requiring the ci.yml checks once Task 1 lands

## Review Findings Index

Full per-area review reports (2026-07-19) contain the complete finding list with file:line references; the auto plan-review (same date) verified the plan's references against the codebase. Findings addressed by this plan: CI-never-runs, Deploy-Demo failures, WASM persistence loss (C1), float truncation (H1), dead rate-limit info (H2), generator silent skips (H3/M6), unrecovered WASM panics (H4), multi-tab clobbering (H5, lock-only scope), export scoping/truncation (H1-backend), webhook context cancellation (H2-backend), forged webhook events (H3-backend), SSE timeouts (H4-backend), training-load range dependence (M1), mixed-offset timestamps (M4), rate-limiter over-waits (M3), SaveStream size-cap spoofing (M2, cap only), calendar-summary zeros (M4-wasm), init-sleep race (M3-wasm), single-listener event bridge, notifications envelope bypass, silent settings failures, calendar bucketing, heatmap country no-op, pace `:60` rounding, error-boundary/404/code-splitting gaps, E2E hard waits, coverage gaps, docs drift. Deferred (documented, not scheduled): full API authentication/Host-allowlist hardening (M5 — needs a deployment-model decision), storage-layer ctx propagation + WAL pragmas (L7), multi-athlete schema scoping (L11), second-tab read-only UX, web-worker migration for the WASM runtime (M2-wasm — docs corrected instead in Task 25).
