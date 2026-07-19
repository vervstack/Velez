# VPN client key (Headscale auth key) storage in task context

`ConnectServiceToVpnTaskPayload.client_key` (`api/grpc/tasks.proto`) holds the
Headscale client auth key issued for a service's VPN sidecar. It is persisted
as a plain string, verbatim, inside the `velez.tasks.context` JSON column via
the durable jobs engine (`internal/jobs`) - the same way every other task
field is persisted for checkpoint/resume.

## Decision

Keep it as a plain string. Do not add field-level encryption or omit it from
the persisted context.

**Why:** the old in-memory pipeline (`do_connect_service_to_vpn.go`) already
held this key in a plain Go closure variable for the pipeline run's lifetime,
with no encryption at rest. Persisting it in the task context JSON is no
worse than that status quo. Adding encryption-at-rest for this one field
would be a bigger architectural change than the migration this decision was
made for (`ConnectServiceToVpn` pipeline -> jobs, see
`docs/jobs_migration.md`), and out of scope for a behavior-preserving
migration.

**Trade-off accepted:** anyone with read access to the `velez.tasks` table
(or a `WatchTask`/`GetTaskByEntityAction` response before the task reaches
`DONE`) can read a live Headscale auth key in the clear. Headscale auth keys
are reusable-namespace-scoped and revocable server-side, which bounds the
blast radius of a leaked row.

## If this needs to change later

Revisit if:
- The `tasks`/`jobs` Postgres tables gain a broader set of readers (e.g. a
  general-purpose admin UI over raw task rows) beyond the task's own
  `WatchTask` caller.
- Any other job payload starts carrying credentials/secrets, at which point
  a shared field-level encryption helper for `TaskContext` (rather than a
  one-off for this field) would be the right unit of work - ask before
  building it, per this repo's "no architectural decisions without asking"
  rule (`CLAUDE.md`).
