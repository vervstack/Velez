# Developer-profile coverage audit

Walking the repo file by file, mapping **every line** to one of:

- **L** — enforced by an off-the-shelf linter (rule + linter name)
- **C** — candidate for a custom check (callfence-class) — no off-the-shelf linter exists
- **F** — structurally forced by the language / build (no rule needed)
- **P** — deliberate personal preference, not mechanically checkable, recorded here so review knows it is intentional

If a line is blank, it is blank for a reason. If it is a one-liner, it is meant to be.

---

## Config changes made during the walk

| # | Change | Reason |
|---|---|---|
| 1 | `.golangci.yaml` govet: `enable: [shadow]` | Profile: never shadow variables. Repo already has 0 shadow hits — locks the property in. |
| 2 | Removed forbidigo exclusion for `internal/io/std_printer.go` | Path does not exist in Velez; no `fmt.Print*` in non-test code. Leftover from the *zpotify* config. |
| 3 | `pkg/web/ZpotifyUI/node_modules` → `pkg/web/Velez-UI/node_modules` in `linters.exclusions.paths` | Same zpotify leftover; the `formatters` section already had it right. |
| 4 | `.golangci.yaml` `formatters.settings.goimports.local-prefixes: [go.vervstack.ru/Velez]` + `golangci-lint fmt ./...` (60 files, mechanical import regroup) | Item A. Enforces the 3 import groups (stdlib / all-external incl. `go.redsock.ru` + vervstack libs / this module). Surfaced 58 files that interleaved own-module imports with external ones — nothing caught it before. |
| 5 | `internal/app/custom.go`: `rerrors.SetSeparator(':')` added as first line of `Custom.Init` | Item F. 30-go wants it at startup; it was called nowhere. |
| 6 | `internal/app/custom.go:325`: extracted `listReq := &velez_api.ListSmerds_Request{}` | Item G. Inline struct-literal-in-call, banned by project `CLAUDE.md`. |
| 7 | Deleted the `depguard` settings block | Item B. Dead (never in `enable:`) + all-stale zpotify text/paths. |

### Profile (`~/.claude/rules/`) changes

- `30-go.md`: constants → Go-idiomatic `camelCase`/`PascalCase`, not `SCREAMING_SNAKE_CASE` (item I).
- `30-go.md`: `SetSeparator` reworded "in `main()`" → "at startup / app bootstrap".
- `30-go.md` + `20-architecture.md`: import grouping restated as **exactly 3** groups; the 4-step dep list is dep-*selection* order, not grouping.
- `_ledger.md`: dated entry + Settled-table row updated.

## Open items (decided later)

| # | Item | Status |
|---|---|---|
| A | 3 import groups (stdlib / all-external / this-module) not automated. | **DONE** — `goimports.local-prefixes` + `golangci-lint fmt` (config change #4). |
| B | Dead + stale `depguard` block. | **DONE** — deleted (change #7). |
| C | `enable:` list has cosmetic duplicates: `govet` (x2), `dupword` (x2). golangci dedups, harmless. | **DONE** — removed the out-of-alphabetical-order duplicate of each (`govet` line 69, `dupword` line 78), kept the alphabetically-placed one. `golangci-lint config verify` clean. |
| D | No linter enforces the **positive** package-naming preference (snake_case OK, mashed `allasoneword` discouraged). ST1003/revive (which would *complain* about snake_case) are correctly off. Enforcing the preference = custom linter. | Review-only for now. Do not enable revive/ST1003. Also keep them off so they don't fight `Id`/`Api`/`Uuid` casing. |
| E | `decorder` is in `enable:` but **has no settings block → it enforces nothing**. Enabling `dec-order` → ~40+ violations repo-wide. `disable-init-func-first-check` is moot (`gochecknoinits` already bans `init`). | **SKIP** — not a no-brainer. Declaration order stays: `funcorder` (methods) + review-only for free-function order. `decorder` left inert in the list. |
| F | `rerrors.SetSeparator(':')` was called nowhere. | **DONE** — added to `Custom.Init` (change #5). Profile reworded. |
| G | Inline struct-literal-in-call at `custom.go`. | **DONE** — extracted to `listReq` (change #6). Still a custom-check candidate for the rest of the repo. |
| H | `rerrors.Wrap` message convention ("`error <verb>ing …`", lowercase) — **unenforced**. Live violation: `internal/config/load.go:115` `"Error parsing servers to config"` (capital). **Root cause is in verv**, not Velez: `~/verv/verv/plugins/project/go_project/patterns/generators/config_generators/gen_servers.go:21` hard-codes `ErrorMessage: "Error parsing servers to config"` — its sibling `gen_data_sources.go:22` correctly uses lowercase. | Fix at source: `gen_servers.go:21` → `"error parsing servers to config"`. Custom-check candidate remains (regex 2nd arg of `rerrors.Wrap`/`New` must start `[a-z]`). |
| J | `internal/config/load.go` carries `DO NOT EDIT` but looked hand-maintained. **Investigated:** it is **stale generated output**, not a real divergence. The thread-safe `Init`/`Load` + `sync.Once` pattern (added to Velez in `bb28feeb`, 2026-07-26, during the e2e-flaky firefight) **is now in verv's template** (`autoload.go.pattern`) verbatim, comments and all. Velez hasn't re-run codegen because the current template needs matreshka `v1.0.100`'s `ReadConfig(...ReadOption)` API. **`load_test.go`** is a hand-written test of code that should be 100 % generated (violates 32-go / 50-posture). | **Own planned task — bigger than a bump.** See J-blocker below. |
| J-blocker | Bumping Velez matreshka `v1.0.94 → v1.0.100` is a **migration**, not a version bump: `v1.0.100` **deleted the public package `go.vervstack.ru/matreshka/pkg/matreshka_api`** (the generated gRPC client/server stubs) — moved to consumer-side `moti g` generation. **14 non-test Velez files import it** (`internal/cluster/disabled.go` implements the whole `MatreshkaBeAPIServer`; `internal/service/service_manager/configurator/*` ×6; `internal/domain/configuration.go` — itself a 31-go-layering violation; clients, jobs, transport). Velez **already** generates its own copy at `internal/api/clients/matreshka/pkg/matreshka_api/`, so the migration is *probably* "repoint the import path ×14 + fix API drift + `moti g` + rebuild" — but that's a cross-layer change needing its own plan and review (Velez `CLAUDE.md`: "always suggest and ask"). `go get` also pulled `rerrors v0.0.7 → v0.0.8`. **Reverted** — `go.mod`/`go.sum` restored, build clean. | Schedule as a dedicated task: matreshka migration → then `verv project tidy` → then `load_test.go` removal → then "other bumps". |
| K | Test naming: profile 32-go says `TestThing_Scenario`; Velez uses `Test_Thing_Scenario` (leading underscore) pervasively — `Test_Lifecycle`, `Test_ClusterMode_*`, `Test_Load_Concurrent_NoRace` — and `.golangci.yaml` even keys an exclusion on `text: "Test_"`. | **DONE** — reworded globally in `~/.claude/rules/32-go-testing.md` to `Test_Thing_Scenario` (Velez practice was consistent enough, and pervasive enough, to be the correct rule rather than a local exception). `_ledger.md` updated. |
| I | Const naming: profile said `SCREAMING_SNAKE_CASE`, code is Go-idiomatic camelCase. | **DONE** — profile reworded to camel/Pascal. Code was already correct. |
| L | `internal/domain/` violates 31-go-layering's domain-purity rule pervasively — imports of generated proto (`internal/api/server/velez_api`, `internal/api/clients/matreshka/pkg/matreshka_api`), the sqlc-generated repository package `internal/storage/postgres/generated/deployments_queries`, and the third-party Docker SDK (`github.com/docker/docker/api/types/image`). Confirmed offenders: `configuration.go` (velez_api, matreshka_api — the matreshka_api half is the same instance already named under J-blocker), `deployments.go` (`deployments_queries.VelezDeploymentStatus` — domain depends on repository-*generated* types, upside-down relative to the layering diagram), `pg_instance.go` (velez_api, for `PgInstanceIsolation`), `pipeline_request.go` (velez_api — `LaunchSmerd` embeds `*velez_api.CreateSmerd_Request` **by pointer**, so every field/method of the generated message is promoted), `plugins.go` (velez_api, two enum types), `service.go` (velez_api, for `DeploymentStatus`; also imports `go.redsock.ru/toolbox` for `rtb.Optional[string]` — borderline, a generic utility container rather than sql/proto/framework, flagging rather than calling it a violation outright), `labels/verv_labels.go` (docker SDK, in `IsMatreshkaImage`). Clean by contrast: `common.go`, `graph.go`, `image.go`, `metrics.go`, `network.go`, `node.go`, `registry.go`, `secret.go`, `service_labels.go`, `upgrade.go`, `vcn.go`, `volume.go`, `labels/compose_labels.go`, and the whole `vervonomicon/` subpackage (genuinely pure — its own doc-comment says so, and it holds up). | **DONE** — Velez-local override, not a global profile change: recorded in `CLAUDE.md` ("Domain layer override" section) that `internal/domain/` may use generated proto/sqlc types directly. `vervonomicon/` is explicitly carved out and stays pure. No code changed. |
| M | Initialism casing inconsistent *within* `internal/domain/` itself: `environment.go` (`ID`), `registry.go` (`ID`), `pg_instance.go` (`ServiceID`), `metrics.go` (`CPUPercent`) use all-caps, while `deployments.go`/`node.go` in the same package correctly use `Id`/`ServiceId`/`CpuPercent`. Same class of gap as the `Api`/`API` note already recorded for `custom.go` under item D — unenforced, since revive/ST1003 are deliberately kept off. | **DONE** — renamed `Environment.ID`→`Id`, `Registry.ID`→`Id`, `PgInstance.ServiceID`→`ServiceId`, `ContainerStats`/`ServiceMetrics.CPUPercent`→`CpuPercent` (plus their req/view sibling structs) and every call site across `internal/storage/`, `internal/service/`, `internal/transport/`, `internal/clients/`, and their tests. Verified with `go build`, `go vet ./...`, `golangci-lint run ./...`, `go test ./internal/...` — all clean. |
| N | `internal/domain/secret.go`'s `ErrInvalidSecretRef` is declared via stdlib `errors.New`, not `rerrors.New` — violates 30-go's "everything goes through rerrors" (the `errors.Is`-is-fine exception already noted for `custom.go` covers `errors.Is`, not `errors.New`). `err113` doesn't catch it: a package-level sentinel declaration is exactly the pattern err113 is designed to allow through. | **DONE** — switched to `rerrors.New("invalid secret ref")`. |

## Callfence status (declaration order)

| Concern | Covered? |
|---|---|
| Constructor sits with its struct; exported methods before unexported (per type) | ✅ `funcorder` (enabled, passing) |
| `const`/`var`/`import` grouped into single blocks | ✅ `grouper` (enabled) |
| File-global `const → type → var → func` order | ❌ `decorder` enabled but unconfigured (open item E) |
| Free function defined **in call order / define-before-use** (the original callfence goal) | ❌ no linter exists — custom only. e.g. `custom.go` defines `parseLogLevel` / `seedOccupiedPorts` / `smerdsDropper` near the bottom, all first called near the top. Deliberate style choice — record as **P** unless callfence gets built. |

---

## `cmd/service/main.go` — DONE

19 lines. `golangci-lint run ./cmd/...` → 0 issues.

| Lines | Code | Class | Detail |
|---|---|---|---|
| 1 | `package main` | F | forced |
| 2 | blank | L | `gofmt` / `wsl_v5` |
| 3–7 | import block | L | `grouper` — single grouped block required. 2 tiers here (3rd-party, blank, own-module) matches 30-go. Tier *split* itself → open item A. |
| 5 | blank line inside imports | L | `goimports` keeps it; separates third-party from own-module (30-go) |
| 9 | `func main() {` | F/P | named func decl (profile: named only) — not checkable, `main` is forced anyway |
| 10 | `a, err := app.New()` | P | one-symbol var `a` for a function this small is deliberate — clear and easy. `varnamelen` intentionally off. |
| 11–13 | `if err != nil { log.Fatal()... }` | L | `noinlineerr` — err is not checked inline. 00-core, fully enforced (linter **and** CLAUDE.md). |
| 12, 17 | `log.Fatal().Err(err).Send()` | L + P | `zerologlint`, `loggercheck` — typed chain, no `-f` call (30-go). `.Send()` over `.Msg("")` is preference — reads best. |
| 14 | blank between the two err blocks | L | `wsl_v5 default: all` — blank-line placement is enforced |
| 15 | `err = a.Start()` (reuse, not `:=`) | L | `wastedassign`, `ineffassign`, `govet shadow` — no shadowing |
| 16–18 | second err check | L | `noinlineerr`, `wsl_v5` |
| 19 | `}` | F | |

**Deviation resolved:** profile 30-go says `rerrors.SetSeparator(':')` is called "in `main()`". Decision: it belongs in the **app bootstrap** (`internal/app`), not `main.go` — `main` stays minimal and carries no specifics. Profile rule to be reworded: *"called once during startup (app bootstrap), not necessarily in `main()` itself."* → **it is currently called nowhere** (open item F).

---

## `internal/app/` — DONE

`golangci-lint run ./internal/app/...` → 0 issues.

### `app.go`, `config.go`, `server.go` — generated, OUT OF SCOPE

Header: `// Code generated by RedSock CLI. DO NOT EDIT.` → read-only (00-core), regenerated by `verv project tidy`. golangci-lint auto-detects the marker and suppresses most issues. Not audited line-by-line. Known deviations that are **acceptable because the profile exempts generated code**:

- `app.go:26` `MASTER net.Listener` — all-caps initialism (30-go bans; generated).
- `app.go:33` `func New() (app App, err error)` — returns a value, not `*App` (30-go wants concrete pointer; generated).
- `app.go:66–84` — `x := func() … {}()` IIFE-assignment pattern (30-go bans `var f = func(){}`; generated).
- `app.go:21–22` `Ctx context.Context` stored in struct (30-go: "never stored in a struct") — carries a `//nolint:containedctx` with a written reason. Generated.
- `app.go:9`, `config.go`/`server.go` import `github.com/rs/zerolog/log` directly — what the dead depguard rule (item B) would target.

### `custom.go` — IN SCOPE (header says `DO EDIT`, so golangci lints it fully)

381 lines, lint-clean. Line classes:

| Area | Class | Detail |
|---|---|---|
| L43–47 `const (…)` — `time.Second * N` extracted to named consts | L | `mnd` forces the extraction; `grouper` forces the single block. ✅ magic-number → token working as intended. |
| const **names** camelCase (`defaultTaskWorkerTimeout`) | I | 30-go says SCREAMING_SNAKE — open item I |
| L49–74 `Custom` struct, field blank-line groupings + `//` group headers | L | every blank line here is `wsl_v5`-approved; grouping comments are **P** (deliberate sectioning) |
| L61–66 `ApiGrpcImpl`, `VpnApiImpl`, … | P | `Api` not `API` (30-go) — not enforced (revive off, item D), review-only |
| L76–287 `*Custom` methods then free funcs | L | order enforced by `funcorder` (constructor/struct-method + exported-before-unexported) |
| every `err = f()` on its own line, `if err != nil` next | L | `noinlineerr` |
| every `return` preceded by blank line | L | `nlreturn` |
| all errors via `rerrors.New` / `rerrors.Wrap`, lowercase "error …ing" messages | L (no dynamic errs) + H (message format unenforced) | `err113` + `wrapcheck` cover the *what*; the message wording is item H |
| L166 `errors.Is(err, cmux.ErrServerClosed)` | F | stdlib `errors.Is` is allowed (not `errors.New`) |
| L84–100, 289–294 multi-line `//` comments explaining non-obvious wiring / ordering constraints | P | genuine external-gotcha comments (00-core permits) — resolver-must-exist-before-seed, daemon-wide port view, etc. |
| L162–188 `errgroup` + `g.Go(func() error {…})` anonymous closures | F/P | idiomatic errgroup; 30-go's "no `var f = func(){}`" targets assignments, not API callbacks |
| L316–367 `smerdsDropper` returns a `func() error` closure | P | factory-returns-closure — matches 30-go's `closer.Add(...)` cleanup idiom; not a banned assignment |
| L139, L155 `go c.DeployWatcher.Start(a.Ctx)` — bare goroutine, return value dropped | — | profile has **no rule yet** (concurrency/goroutines is a pending 30-go item); `errcheck` does not inspect `go` statements |
| L235 `log.Info().Bool("shutDownOnExit", …)` field key camelCase vs `"count"`/`"uuid"` snake elsewhere | P | inconsistent, unenforced (`loggercheck` checks pairing not key casing) |
| **L323** `ListSmerds(ctx, &velez_api.ListSmerds_Request{})` — inline struct literal in call | **G** | violates project `CLAUDE.md`; unenforced; fix + custom-check candidate |
| L321 `ctx := context.Background()` inside the shutdown closure | P | shutdown hook has no ctx to thread — deliberate |

---

## `internal/config/` — DONE

`golangci-lint run ./internal/config/...` → 0 issues.

### Generated, OUT OF SCOPE — `environment.go`, `data_sources.go`, `keys.go`, `servers.go`

All `// Code generated by RedSock CLI. DO NOT EDIT.`, regenerated by `verv project tidy`. golangci skips them. Acceptable deviations (generated exempt):

- `environment.go`: `CPUDefault`, `DisableAPISecurity`, `MakoshURL`, `MatreshkaImage`, `RAMMbDefault`, `VpnServerURL` — `CPU`/`API`/`URL`/`RAM` all-caps initialisms (30-go bans). Driven by matreshka env keys.
- `environment.go:36,42` two separate `const (…)` blocks — `grouper` would want one; generated-exempt.
- `servers.go:11` `MASTER *server.Server` — all-caps.
- Exported enum consts (`LogFormatText`, `ResourceGrpcMakosh`, …) are PascalCase ✅ (matches reworded 30-go).

### `load.go` — HAND-MAINTAINED but wrongly marked generated → **item J**

Currently lint-excluded. If un-excluded it needs a `grouper` pass (two `var` decls: `ErrAlreadyLoaded` L17 + the `var (…)` block L30) and the line-115 wrap message (**item H**). The `flag.StringVar` / `os.Stat` / `sync.Once` logic itself follows 30-go (2-line err checks, `nlreturn`, grouped consts).

### `load_test.go` — IN SCOPE, clean

| Point | Class | Detail |
|---|---|---|
| `package config` (white-box) for exported `Load` | P | 32-go allows it; black-box `config_test` would be tidier |
| `Test_Load_Concurrent_NoRace` name | P | was flagged as item K (profile vs. Velez mismatch); item K resolved by rewording the profile, so this is now just the documented convention — still unenforced by any linter |
| `require.NoError` | L | 32-go: `require` is the default |
| straight-line, not table (1 case) | L/P | 32-go: don't pre-emptively table |
| race-freedom unit test for `Load` | — | 50-testing-posture: fragile-logic (concurrency) is what unit tests are *for* ✅ |
| `for range 20` — the `20` | L | passes `mnd`; `intrange` likes the form |
| no `t.Parallel()` | — | touches no Docker/global namespace — CLAUDE.md only requires it there |

---

## `internal/domain/` — DONE

`golangci-lint run ./internal/domain/...` → 0 issues. This is a pure-data-type package (28 files,
~1340 lines across the top-level package plus the `labels/` and `vervonomicon/` subpackages) — its
whole job under 31-go-layering is to *not* import anything but stdlib and primitive-type libraries.
That's exactly where it deviates most: see items **L**, **M**, **N** above, all newly opened by this
walk.

### Top-level `internal/domain/*.go`

| File | Lines | Class | Detail |
|---|---|---|---|
| `common.go` | 6 | F | `Paging{Limit, Offset uint64}` — pure, no imports |
| `configuration.go` | 37 | **C (item L)** | imports `matreshka_api` and `velez_api` (both generated/proto) — `ConfigMeta.ConfType`/`.Format` are foreign enum types living directly in a domain struct |
| `deployments.go` | 35 | F / **C (item L)** | `Id`/`ServiceId`/`SpecId`/`NodeId` casing correct (F, matches profile) — but `Status deployments_queries.VelezDeploymentStatus` imports the **sqlc-generated repository package**, the one direction 31-go-layering never allows |
| `environment.go` | 65 | F / **P (item M)** | genuinely excellent doc-comments explaining `Suffix` vs Docker host semantics (00-core's "genuine external gotcha" bar, easily) — undercut by `ID int64` (should be `Id`) |
| `graph.go` | 33 | F | pure, stdlib `time` only |
| `image.go` | 10 | F | pure, no imports |
| `metrics.go` | 27 | F / **P (item M)** | `CPUPercent float64` — all-caps initialism, inconsistent with `node.go`'s `CpuPercent` two files over |
| `network.go` | 8 | F | pure, no imports |
| `node.go` | 27 | F | pure, stdlib `time` only; `Id`/`CpuPercent`/`MemPercent` all correctly cased — the positive counter-example to `environment.go`/`metrics.go` |
| `pg_instance.go` | 116 | F / **C (item L)** / **P (item M)** | best-commented file in the package (explains the whole PGaaS "never duplicated" invariant) — but `Isolation velez_api.PgInstanceIsolation` is a proto enum in a domain struct, and `ServiceID` should be `ServiceId` |
| `pipeline_request.go` | 24 | **C (item L)** | `LaunchSmerd` embeds `*velez_api.CreateSmerd_Request` by pointer — the sharpest purity violation in the package, since embedding (vs. a named field) promotes the entire generated message's surface into the domain type |
| `plugins.go` | 15 | **C (item L)** | `pb.VervPlugin_State`, `pb.DeploymentStatus` — two more proto enums straight in a domain struct |
| `registry.go` | 48 | F / **P (item M)** | `Url string` correctly cased (matches profile's example exactly) — but `ID int64` should be `Id` |
| `secret.go` | 54 | F / **C (item N)** | otherwise a model domain file: pure (`errors`, `slices`, `strings` only), named sentinel, real parse/round-trip logic worth the unit test it has — the one deduction is `errors.New` instead of `rerrors.New` for `ErrInvalidSecretRef` |
| `secret_test.go` | 101 | L | `package domain_test` (black-box, the tidier option 32-go mentions), table-driven at 4 cases, `require.Error`/`require.ErrorIs`/`require.NoError`/`require.Equal` throughout, plus a dedicated round-trip test (`String` → `ParseSecretRef` → equal) — no notes, this is the reference-quality test file in the package |
| `service_labels.go` | 68 | F | pure (`strings` only); doc-comments earn their keep explaining the derived-never-persisted label scheme |
| `service.go` | 92 | **C (item L)** | `velez_api.DeploymentStatus` (proto) in `Service`; `rtb.Optional[string]` (`go.redsock.ru/toolbox`) in `ListServicesReq` is the one *borderline* case in item L, not a clear-cut violation |
| `upgrade.go` | 6 | F | pure, no imports |
| `vcn.go` | 42 | F | pure; the only file in the package carrying `json:"..."` tags (these types round-trip through `internal/workers` JSON persistence, not proto) |
| `vervonomicon_deploy.go` | 24 | F | imports only `internal/domain/vervonomicon` — in-layer, not a purity issue |
| `volume.go` | 6 | F | pure, no imports |

### `internal/domain/labels/`

| File | Lines | Class | Detail |
|---|---|---|---|
| `compose_labels.go` | 5 | F | pure, one constant |
| `verv_labels.go` | 57 | F / **C (item L)** | mostly a well-commented label-constant catalogue (each constant's Detail explains its own read path, e.g. `PgaasInstanceLabel`) — `IsMatreshkaImage(r *image.InspectResponse) bool` imports `github.com/docker/docker/api/types/image` directly, the one place *any* subpackage of `domain` reaches for a third-party SDK type |

### `internal/domain/vervonomicon/`

`golangci-lint run ./internal/domain/vervonomicon/...` → 0 issues.

| File | Lines | Class | Detail |
|---|---|---|---|
| `auth.go` | 13 | F | pure, no imports; `yaml:"...,omitempty"` tags only |
| `box.go` | 12 | F | pure; `Cpu float64` — correctly cased, contrast with top-level `metrics.go`'s `CPUPercent` |
| `deployment.go` | 85 | F | pure; every field commented with its CreateSmerd.Request mapping or a genuine "why" (e.g. `Image` only set for repo/client-push sources) |
| `descriptor.go` | 75 | F | pure; package doc-comment explicitly states the purity intent ("this package only mirrors the YAML shape") — and every other file in the subpackage backs that claim up |
| `ingress.go` | 37 | F | pure |
| `resource.go` | 57 | F | pure; `ResourceSizing` embeds `Sizing` via `yaml:",inline"` — clean composition, no purity concern |

This subpackage is the strongest positive result of the whole `internal/domain/` walk: zero imports
beyond stdlib (in fact zero imports at all) across six files, closed-vocabulary string enums done
right (`SourceKind`, `Protocol`, `RestartPolicy`, `Ssl`, `Isolation` — all with `_UNSPECIFIED`-style
completeness via doc comments, though proto's mandatory zero-value convention doesn't apply to a
plain Go string type), and comments that consistently clear the "genuine external gotcha" bar
instead of restating the code.
