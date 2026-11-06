# Releases and E2E in CI

- Every master push: unit tests, then an rc release - `vX.Y.Z+1-rc` image, git tag and GitHub Release (`.github/workflows/release.yaml`, version from `.github/scripts/next-version.sh`). `:latest` does not move.
- Minor release: run the `release-minor` workflow manually on master. It runs the full e2e, then retags the HEAD commit's rc image as `vX.Y+1.0` and `:latest`, tags and creates the GitHub Release. HEAD must already have its rc tag.
- Trello: a commit names its card with a `https://trello.com/c/<shortLink>` URL in the body. The rc push moves each linked card still
  before `review/test` into `review/test` (Artel tract `velez-rc-to-test`); the minor release moves linked cards that are in
  `review/test` with every checklist item checked (cards without a checklist stay) into the open version list `vX.Y.0` on the
  Verv board, archives it and creates the next one (tract `velez-minor-release`). Both are called by `.github/scripts/trello-notify.sh`
  with the secrets `TRACT_RC_*` / `TRACT_RELEASE_*`; the tract webhook takes its token in the `X-Gitlab-Token` header.

## Running e2e locally

Outside the minor release, e2e does not run in CI. Run e2e locally, by group, never the whole suite while iterating:

- Run the groups your change touches: `make e2e-group GROUP=<name>` (group list: `tests/e2e/groups.yaml`;
  tests outside every listed group are in `rest`). A change to `internal/jobs/` runners, DinD or BuildKit
  code means `runners`; to the statefull/enable flow means `statefull` and `enable_statefull`; and so on.
- A filtered `go run ./tests/e2egroups run --group <name> --filter '<regex>'` narrows a group to the tests
  you touched.
- Run the full `make e2e` (about 12 minutes; `make e2e MAX_GROUPS=N` limits concurrent groups) only before
  handing back a large cross-cutting change.
