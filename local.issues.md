 ---
Code Issues Noticed During Research

Not blocking the tests, but worth a follow-up fix:

1. internal/pipelines/steps/healthcheck.go:62 — Logic bug: if cont.State.Health == nil { continue } prevents checking
   State.Status == "running" for containers
   without a native Docker healthcheck. When no Command is set (no Docker-native healthcheck is configured on the
   container), all retries are skipped and
   errRetriesExhausts is always returned. Fix: when Health == nil, treat Status == "running" as success.
2. internal/service/service_manager/container_manager/inspect.go:36 — Potential panic: imageInfo.RepoTags[:1] panics
   when RepoTags is empty (locally built or
   dangling image). Fix: if len(imageInfo.RepoTags) > 0 { smerd.ImageName = imageInfo.RepoTags[0] }.
3. internal/service/service_manager/container_manager/smerd_list.go:17 — Mutates the incoming request's Name in-place
   via *req.Name = strings.ToLower(...). Fix:
   compare lowercased copy, don't write back.
4. internal/service/service_manager/container_manager/smerds_drop.go:12 — Variable named uuid iterates over both UUIDs
   and names. Rename to target.
5. Code duplication: tests/helper*.go ≈ tests/e2e/helper*.go (only difference is e2e_ prefix). Low priority but worth
   consolidating.

 ---