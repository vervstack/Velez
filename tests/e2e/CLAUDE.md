# tests/e2e — testing rules

Scope: every file in this directory (Claude Code auto-loads this file whenever a
`tests/e2e/*.go` file is read or edited — no need to repeat it in the root CLAUDE.md).

## Two tiers of coverage

- **Always-on smoke test** — `suite_container_runtime_test.go`'s `Test_ContainerRuntime_Matrix`.
  No `//go:build` tag, so it runs in every `go test ./tests/e2e/...` and every CI invocation.
  Keep it fast and narrow: one real CreateSmerd → real Docker container round trip per matrix
  cell, nothing more.
- **Everything else** — gated behind `//go:build e2e_full`. Deeper suites (network isolation,
  list-scoping, upgrade, statefull, cluster, vpn, ...) live here. Run manually with
  `go test -tags e2e_full ./tests/e2e/...` before handing back work that touches core logic (see
  root `CLAUDE.local.md`).

A new e2e test file defaults to `e2e_full` unless it is cheap enough to run on every commit —
ask before adding a second always-on test.

## The Plane matrix (`matrix_test.go`)

`Plane` names one cell of a package-wide fixture-shape matrix, over four axes:

| Axis | Type | Wired today | Not wired (tracked: https://trello.com/c/otriSswo) |
| --- | --- | --- | --- |
| State backend | `PlaneMode` | `ModeSingleNode`; `ModeCluster` wired, via `ClusterPlanes` (see below) | — |
| Container-runtime backend | `PlaneBackend` | `BackendDocker` | — |
| Environment separation | `EnvSeparationWay` | `SeparationLabelBased` (one Docker engine, envs kept apart by name/label/suffix) | `SeparationDedicatedEngine` (one Docker daemon per environment) |
| App execution | `RunningMode` | `RunningModeBinary` (in-process, bufconn) | `RunningModeContainer` (real Velez container over the network) |

`Planes` is the package-wide list of cells; `Planes[0]` is the only one wired to a real fixture
today. `Plane.Name()` is *generated* from the four axis values (`strings.Join(..., "/")`) — never
add a separate hand-set display-name field, it can silently drift from what the cell actually is.

**Adopting the matrix in a suite**: iterate `Planes` (or a suite-local subset) and build the
fixture through `plane.NewEnvironment(t, opts...)`, never `NewEnvironment(t, opts...)` directly.
`Plane.NewEnvironment` guards each axis and calls `t.Skipf(...)` — not `t.Fatalf` — for any
combination not wired to a real fixture yet, so an unimplemented cell is reported as skipped, not
a false failure, and the skip message links to the tracking ticket. Adding a `Planes` row ahead of
its fixture support is safe by construction: every suite iterating it just skips the new cell
until `NewEnvironment` is taught to build it.

When a new axis value becomes real (e.g. `SeparationDedicatedEngine` gets implemented), wire it
once in `Plane.NewEnvironment` and every adopting suite picks it up — that's the entire point of
routing through `Plane` instead of each suite hand-rolling its own fixture setup.

## Adopting the matrix in a `testify/suite` suite

Every suite in this package (bar `suite_container_runtime_test.go`, which predates this and hand-
rolls its own function-per-axis tree) already uses `testify/suite`. The matrix-aware shape for one
of these is: give the suite a `plane Plane` field, build its fixture through `s.plane.NewEnvironment`
(wherever the suite already builds it — `SetupSuite`, `SetupTest`, or inline per test method) instead
of `Planes[0].NewEnvironment`/`NewEnvironment`, and drive it with `RunPlaneSuite` instead of a bare
`suite.Run`:

```go
type ControlPlaneSuite struct {
	suite.Suite

	plane Plane
}

func (s *ControlPlaneSuite) Test_ListEnvironments_WithLocalStateConfig() {
	env := s.plane.NewEnvironment(s.T(), ...)
	...
}

func Test_ControlPlane(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ControlPlaneSuite{plane: plane}
	})
}
```

`RunPlaneSuite` (`matrix_test.go`) runs `newSuite` once per `Plane` in the slice passed to it, each
as its own `t.Run(plane.Name(), ...)` subtest. It adds no skip logic of its own — `Plane.NewEnvironment`
is still the only place that decides a cell isn't wired yet, so passing the full `Planes` (not just
`Planes[0]`) costs nothing today and means the suite needs no further changes when a new cell lights
up. A suite whose test methods are genuinely cluster/matreshka-shaped (`suite_api_deploy_test.go`'s
`Test_ClusterMode_*`, `suite_hello_world_cluster_test.go`, `enableStatefullPgUnderDind`) stays on
`ClusterPlanes[0]` directly rather than looping through `RunPlaneSuite`. `ClusterPlanes` is a
separate package-level var from `Planes`, deliberately: if its cell lived in `Planes` instead,
every other suite adopting `Planes` via `RunPlaneSuite` would pick it up too and pay for a real
DinD cluster fixture it has no cluster-shaped test logic to exercise. `ClusterPlanes[0]` itself
does not enable matreshka - it's a routing point, not a bundle of options - so a suite that needs
`WithMatreshka()` still passes it explicitly. See the doc comment on `ClusterPlanes` in
`matrix_test.go` for the full reasoning.

## Environment axis: name IS the wire value

`velez_api.CreateSmerd_Request.Environment` (and `ListSmerds_Request.Environment`, etc.) *is* the
environment's name — there is no separate `suffix` field to plumb through a test. Concretely:

- `environments.DefaultEnvironmentName` (`"PROD"`) is the real production sentinel for the node's
  own/default environment. `internal/storage/environments/static.go`'s `NewStatic` always seeds a
  row under that exact name, so sending it explicitly resolves identically to sending `""`. Use
  the named constant in test tables instead of `""` or a repeated `"PROD"` literal — it reads the
  same as any other environment case instead of needing a magic-empty-string special case.
- A non-default environment's suffix **is its own name** (`NewStatic` seeds `(name, name)` for
  every other registered environment). So a table of environment cases needs exactly one field,
  not `{name, environment}` or `{environment, suffix}` — collapsing those was a real PR review
  fix (#70), not a hypothetical.
- **`ContainerSuffix` (`WithContainerSuffix`) is a different axis**: it's the *default/PROD*
  environment's suffix for this one fixture, generated from `t.Name()`
  (`dockerSafeToken(t.Name())` in `suite_container_runtime_test.go`) so two parallel tests
  deploying to PROD never collide in Docker's global container namespace. It exists only because
  one Docker engine can host multiple environments (`SeparationLabelBased`); it is meaningless
  once `SeparationDedicatedEngine` is wired — a dedicated engine already isolates by construction
  and gets no suffix.
- Register every non-default environment a fixture needs via `WithEnvironments([]string)` —
  derive that list from the same table driving the test cases (see
  `containerRuntimeExtraEnvironments()`) instead of hand-listing it a second time, so a new table
  row can't drift out of sync with what's actually registered on the fixture.

## Avoid N nested loops in one function

A matrix test (Plane × environment × state-mode × variant, ...) tempts one function with a loop
per axis. Split one function per axis instead — each takes `t.Helper()` + `t.Parallel()` and
calls `t.Run` into the next axis's function
(`Test_ContainerRuntime_Matrix` → `runContainerRuntimePlane` → `runContainerRuntimeEnvironmentCase`
→ `runContainerRuntimeStateModeCase` → `runContainerRuntimeCase` is the reference shape). This
keeps every function readable at a glance and the `t.Run` tree still lets you target one cell
directly (`-run Matrix/single-node.../PROD/stateless/hello-world`).

## Fixture gotchas

- `enableStatefullPgUnderDind` (`helper_statefull_test.go`, `e2e_full`) is **not parallel-safe** —
  it `t.Chdir`s to the repo root because `sqldb.RollMigration` resolves `"./migrations"`
  relatively. Never call it from a `t.Parallel()` test.
- `WithMatreshka()` is a single-process singleton shared via `main_test.go`'s `TestMain` — any
  test using it must live in package `tests/e2e`, never a new package (a different package is a
  different OS process and can't see the singleton, reintroducing the container-name collision
  it exists to prevent). See `docs/plans/e2e_flaky_lifecycle_matreshka.md`.
- Every e2e test runs under a DinD harness (`requireDindHarness`) — never against a developer's
  real Docker daemon. `NewEnvironment` enforces this; there's no opt-out.
