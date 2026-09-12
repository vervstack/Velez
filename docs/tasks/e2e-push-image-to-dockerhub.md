# e2e — push locally-built image to Docker Hub on success

**Status:** deferred, not scheduled

## Context

Trello card https://trello.com/c/otriSswo adds two smoke-matrix axes to
`tests/e2e/suite_container_runtime_test.go`: separate-engine (this repo's
worktree) and run-as-container. The run-as-container axis needs a real Velez
image to launch as a container instead of running Velez in-process.

While scoping that axis, running e2e against a locally built-and-tagged image
(`make build-local-container`, tag `velez:e2e`) came up, along with a
follow-on idea: after that e2e run succeeds, push the same image to Docker
Hub as a lightweight release/publish signal.

## Decision

Out of scope for the run-as-container axis itself — building/tagging/reusing
a local image for the test is orthogonal to publishing it. Mixing test
execution with release/CI concerns (registry credentials, tagging scheme,
push failure handling) inside the e2e suite is the wrong layer for it. This
file exists so the idea isn't lost, not to schedule it.

## If picked up later

- Belongs in CI (a workflow step), not in `tests/e2e/...` Go code.
- Needs a tagging scheme (`latest`, git sha, semver?) and Docker Hub
  credentials wired as CI secrets, not local/dev config.
- Should only push on the default branch / a tagged release, never on every
  e2e run (e.g. a PR from a fork should never be able to trigger a push).
- Failure handling: a failed push after green e2e shouldn't fail the e2e job
  itself - it's a separate concern with its own retry/alerting.
